package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awsSQS "github.com/aws/aws-sdk-go-v2/service/sqs"
	sqsTypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"go.uber.org/fx"

	"github.com/FelipeSoft/backend-challenge-go/internal/delivery/queue"
	"github.com/FelipeSoft/backend-challenge-go/internal/platform"
	"github.com/FelipeSoft/backend-challenge-go/internal/wager"
	"github.com/FelipeSoft/backend-challenge-go/internal/wallet"
)

func main() {
	fx.New(
		platform.ConfigModule,
		platform.DatabaseModule,
		wallet.Module,
		wager.Module,
		queue.SQSModule,
		fx.Provide(func() *slog.Logger {
			return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}))
		}),
		fx.Provide(func(logger *slog.Logger) (*awsSQS.Client, error) {
			cfg, err := config.LoadDefaultConfig(context.TODO(),
				config.WithRegion("us-east-1"),
				config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
			)
			if err != nil {
				logger.Error("Falha ao carregar configuração da AWS", "error", err)
				return nil, err
			}
			return awsSQS.NewFromConfig(cfg, func(o *awsSQS.Options) {
				o.BaseEndpoint = aws.String("http://localhost:4566")
			}), nil
		}),
		fx.Provide(func(cfg *platform.Config) string {
			if cfg.SQSQueueURL == "" {
				panic("missing AWS SQS queue URL")
			}
			return cfg.SQSQueueURL
		}),
		fx.Invoke(func(lc fx.Lifecycle, client *awsSQS.Client, consumer *queue.SQSConsumer, logger *slog.Logger) {
			ctx, cancel := context.WithCancel(context.Background())
			lc.Append(fx.Hook{
				OnStart: func(c context.Context) error {
					logger.Info("Ensuring the input and output SQS queue creation (Outbox)...")
					inDlqName := "wager-transactions-dlq.fifo"
					inQueueName := "wager-transactions.fifo"
					inDlqRes, err := client.CreateQueue(c, &awsSQS.CreateQueueInput{
						QueueName: aws.String(inDlqName),
						Attributes: map[string]string{
							"FifoQueue":                 "true",
							"ContentBasedDeduplication": "true",
						},
					})
					if err != nil {
						logger.Error("Error to create input DLQ in SQS", "error", err)
						return err
					}
					inDlqAttrRes, err := client.GetQueueAttributes(c, &awsSQS.GetQueueAttributesInput{
						QueueUrl: inDlqRes.QueueUrl,
						AttributeNames: []sqsTypes.QueueAttributeName{
							sqsTypes.QueueAttributeNameQueueArn,
						},
					})
					if err != nil {
						logger.Error("Error to query input DLQ ARN", "error", err)
						return err
					}
					inDlqArn := inDlqAttrRes.Attributes["QueueArn"]
					inRedrivePolicy := map[string]string{
						"deadLetterTargetArn": inDlqArn,
						"maxReceiveCount":     "5",
					}
					inPolicyBytes, _ := json.Marshal(inRedrivePolicy)
					_, err = client.CreateQueue(c, &awsSQS.CreateQueueInput{
						QueueName: aws.String(inQueueName),
						Attributes: map[string]string{
							"FifoQueue":                 "true",
							"ContentBasedDeduplication": "true",
							"RedrivePolicy":             string(inPolicyBytes),
						},
					})
					if err != nil {
						logger.Error("Error to create main input queue in SQS", "error", err)
						return err
					}
					logger.Info("All SQS queues (input and output) was provisioned successfully. Starting consumer...")
					go func() {
						if err := consumer.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
							logger.Error("SQS Consumer stopped with a critical error", "error", err)
						}
					}()
					return nil
				},
				OnStop: func(c context.Context) error {
					logger.Info("SIGTERM signal received. Stopping SQS consumer with graceful shutdown...")
					cancel()
					stopCtx, stopCancel := context.WithTimeout(c, 10*time.Second)
					defer stopCancel()
					if err := consumer.Stop(stopCtx); err != nil {
						logger.Error("Error to stop SQS consumer", "error", err)
						return err
					}
					logger.Info("SQS consumer stopped successfully.")
					return nil
				},
			})
		}),
	).Run()
}