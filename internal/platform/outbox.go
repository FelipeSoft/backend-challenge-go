package platform

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsSQS "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

type OutboxWorker struct {
	db        *pgxpool.Pool
	sqsClient *awsSQS.Client
	queueURL  string
	logger    *slog.Logger
	batchSize int
	interval  time.Duration
}

type OutboxWorkerParams struct {
	fx.In
	DB        *pgxpool.Pool
	SQSClient *awsSQS.Client
	QueueURL  string `name:"outbox_queue_url"`
	Logger    *slog.Logger
}

func NewOutboxWorker(p OutboxWorkerParams) *OutboxWorker {
	return &OutboxWorker{
		db:        p.DB,
		sqsClient: p.SQSClient,
		queueURL:  p.QueueURL,
		logger:    p.Logger.With("component", "OutboxWorker"),
		batchSize: 10,
		interval:  2 * time.Second,
	}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	w.logger.Info("Outbox worker started successfully.")
	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Outbox worker receiving stop signal.")
			return
		case <-ticker.C:
			if err := w.processBatch(ctx); err != nil {
				w.logger.Error("Error to process outbox batch", "error", err)
			}
		}
	}
}

func (w *OutboxWorker) processBatch(ctx context.Context) error {
	tx, err := w.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("fail to begin outbox transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	query := `
		SELECT id, event_id, event_type, aggregate_id, payload, attempts
		FROM outbox
		WHERE published_at IS NULL 
		  AND next_delivery_at <= NOW()
          AND (locked_until IS NULL OR locked_until <= NOW())
		ORDER BY next_delivery_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED;
	`
	rows, err := tx.Query(ctx, query, w.batchSize)
	if err != nil {
		return fmt.Errorf("fail to query pending events in outbox: %w", err)
	}
	type outboxRecord struct {
		id          string
		eventID     string
		eventType   string
		aggregateID string
		payload     []byte
		attempts    int
	}
	var records []outboxRecord
	for rows.Next() {
		var r outboxRecord
		if err := rows.Scan(&r.id, &r.eventID, &r.eventType, &r.aggregateID, &r.payload, &r.attempts); err != nil {
			rows.Close()
			return fmt.Errorf("fail to scan outbox: %w", err)
		}
		records = append(records, r)
	}
	rows.Close()
	if len(records) == 0 {
		return nil
	}
	var ids []string
	for _, r := range records {
		ids = append(ids, r.id)
	}
	_, err = tx.Exec(ctx, `
		UPDATE outbox 
		SET locked_until = NOW() + INTERVAL '30 seconds' 
		WHERE id = ANY($1)
	`, ids)
	if err != nil {
		return fmt.Errorf("fail to execute a temporary lock in registers: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("fail to confirm lock in outbox: %w", err)
	}
	for _, r := range records {
		err := w.publishEvent(ctx, r.eventID, r.eventType, r.aggregateID, r.payload)
		if err != nil {
			w.logger.Error("Temporary fail to publish event in SQS. Applying backoff...",
				"eventId", r.eventID, "attempt", r.attempts+1, "error", err)
			w.handleFailure(ctx, r.id, r.attempts+1)
			continue
		}
		w.markAsPublished(ctx, r.id)
	}
	return nil
}

func (w *OutboxWorker) publishEvent(ctx context.Context, eventID, eventType, aggregateID string, payload []byte) error {
	messageBody := string(payload)
	input := &awsSQS.SendMessageInput{
		QueueUrl:               aws.String(w.queueURL),
		MessageBody:            aws.String(messageBody),
		MessageGroupId:         aws.String(aggregateID),
		MessageDeduplicationId: aws.String(eventID),
	}
	w.logger.Info("Tentando enviar mensagem para o SQS", "queueURL", w.queueURL, "eventId", eventID)

	_, err := w.sqsClient.SendMessage(ctx, input)
	if err != nil {
		w.logger.Error("FALHA CRUCIAL NO SQS SENDMESSAGE", "error", err)
	}
	return err
}

func (w *OutboxWorker) markAsPublished(ctx context.Context, recordID string) {
	_, err := w.db.Exec(ctx, `
        UPDATE outbox 
        SET published_at = NOW(), locked_until = NULL 
        WHERE id = $1::uuid
    `, recordID)
	if err != nil {
		w.logger.Error("Critical error to mark event as published", "id", recordID, "error", err)
	}
}

func (w *OutboxWorker) handleFailure(ctx context.Context, recordID string, newAttempts int) {
	maxAttempts := newAttempts
	if maxAttempts > 6 {
		maxAttempts = 6
	}
	backoffSeconds := 1 << maxAttempts
	_, err := w.db.Exec(ctx, `
		UPDATE outbox 
		SET attempts = $1, 
		    next_delivery_at = NOW() + MAKE_INTERVAL(secs => $2), 
		    locked_until = NULL 
		WHERE id = $3
	`, newAttempts, backoffSeconds, recordID)
	if err != nil {
		w.logger.Error("Error to update outbox fail", "id", recordID, "error", err)
	}
}
