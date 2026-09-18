package queue

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	wagerdomain "github.com/FelipeSoft/backend-challenge-go/internal/wager/domain"
	"github.com/FelipeSoft/backend-challenge-go/internal/wager/usecase"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"go.uber.org/fx"
)

var SQSModule = fx.Options(
	fx.Provide(NewSQSConsumer),
)

type SQSConsumer struct {
	client                 *sqs.Client
	queueURL               string
	createWagerTransaction *usecase.CreateWagerTransaction
	logger                 *slog.Logger
	isRunning              bool
	consumerName           string
}

func NewSQSConsumer(client *sqs.Client, queueURL string, createWagerTransaction *usecase.CreateWagerTransaction, logger *slog.Logger) *SQSConsumer {
	return &SQSConsumer{
		client:                 client,
		queueURL:               queueURL,
		createWagerTransaction: createWagerTransaction,
		logger:                 logger,
		consumerName:           "wager-transactions-consumer",
	}
}

func (c *SQSConsumer) Start(ctx context.Context) error {
	c.isRunning = true
	c.logger.Info("Iniciando escuta de mensagens SQS", "queue_url", c.queueURL)

	for c.isRunning {
		select {
		case <-ctx.Done():
			c.logger.Info("Contexto cancelado, parando o consumer SQS...")
			return nil
		default:
			output, err := c.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
				QueueUrl:              &c.queueURL,
				MaxNumberOfMessages:   1,
				WaitTimeSeconds:       5,
				AttributeNames:        []types.QueueAttributeName{types.QueueAttributeNameAll},
				MessageAttributeNames: []string{"All"},
			})
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				c.logger.Error("Erro ao receber mensagem do SQS", "error", err)
				time.Sleep(2 * time.Second)
				continue
			}

			for _, msg := range output.Messages {
				if err := c.processMessage(ctx, msg); err != nil {
					c.logger.Error("Falha ao processar mensagem do SQS", "message_id", *msg.MessageId, "error", err)
					continue
				}

				_, err = c.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
					QueueUrl:      &c.queueURL,
					ReceiptHandle: msg.ReceiptHandle,
				})
				if err != nil {
					c.logger.Error("Erro ao deletar mensagem do SQS após sucesso", "message_id", *msg.MessageId, "error", err)
				}
			}
		}
	}
	return nil
}

func (c *SQSConsumer) Stop(ctx context.Context) error {
	c.isRunning = false
	c.logger.Info("Consumer SQS parado com sucesso.")
	return nil
}

func (c *SQSConsumer) processMessage(ctx context.Context, msg types.Message) error {
	if msg.Body == nil {
		return errors.New("mensagem SQS com corpo vazio")
	}

	var envelope SQSMessageEnvelope
	if err := json.Unmarshal([]byte(*msg.Body), &envelope); err != nil {
		c.logger.Error("Erro ao desserializar envelope SQS", "error", err)
		return err
	}

	messageID := envelope.MessageID
	if messageID == "" && msg.MessageId != nil {
		messageID = *msg.MessageId
	}

	c.logger.Info("Processando transação via SQS",
		"message_id", messageID,
		"external_transaction_id", envelope.Data.ExternalTransactionID,
		"provider_id", envelope.Data.ProviderID,
	)

	// 1. Calcular o hash do payload para auditoria/verificação no Inbox
	payloadHash, err := wagerdomain.ComputePayloadHash(
		envelope.Data.ProviderID,
		envelope.Data.ExternalTransactionID,
		envelope.Data.PlayerID,
		envelope.Data.WalletID,
		envelope.Data.Kind,
		envelope.Data.RoundID,
		envelope.Data.GameID,
		envelope.Data.ReferenceExternalTransactionId,
		envelope.Data.Money.Amount(),
		envelope.Data.Money.Currency(),
	)
	if err != nil {
		return err
	}

	// 2. Mapear para o input do UseCase
	idempotencyKey := envelope.Data.IdempotencyKey
	if idempotencyKey == "" {
		idempotencyKey = envelope.Data.ProviderID + ":" + envelope.Data.ExternalTransactionID
	}

	input := usecase.CreateWagerTransactionInput{
		ProviderID:                     envelope.Data.ProviderID,
		ExternalTransactionID:          envelope.Data.ExternalTransactionID,
		PlayerID:                       envelope.Data.PlayerID,
		WalletID:                       envelope.Data.WalletID,
		RoundID:                        envelope.Data.RoundID,
		GameID:                         envelope.Data.GameID,
		Kind:                           envelope.Data.Kind,
		MoneyAmount:                    envelope.Data.Money.String(),
		MoneyCurrency:                  envelope.Data.Money.Currency(),
		IdempotencyKey:                 idempotencyKey,
		ReferenceExternalTransactionId: envelope.Data.ReferenceExternalTransactionId,
		MessageID:                      &messageID,
		ConsumerName:                   &c.consumerName,
		PayloadHash:                    &payloadHash,
	}

	// 3. Executar o UseCase repassando o controle do Inbox (messageID e consumerName)
	// (Recomenda-se expor um método no usecase ou repositório que envolva a gravação do inbox + transação de aposta em uma única transação SQL)
	_, err = c.createWagerTransaction.Execute(ctx, input)
	if err != nil {
		return err
	}

	return nil
}
