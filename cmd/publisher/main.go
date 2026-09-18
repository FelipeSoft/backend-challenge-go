package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/FelipeSoft/backend-challenge-go/internal/platform"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awsSQS "github.com/aws/aws-sdk-go-v2/service/sqs"
	sqsTypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		platform.ConfigModule,
		platform.DatabaseModule,
		fx.Provide(func() *slog.Logger {
			return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}))
		}),
		fx.Provide(func(logger *slog.Logger) (*awsSQS.Client, error) {
			cfg, err := awsConfig.LoadDefaultConfig(context.TODO(),
				awsConfig.WithRegion("us-east-1"),
				awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
			)
			if err != nil {
				logger.Error("Fail to load AWS config", "error", err)
				return nil, err
			}
			return awsSQS.NewFromConfig(cfg, func(o *awsSQS.Options) {
				o.BaseEndpoint = aws.String("http://localhost:4566")
			}), nil
		}),
		fx.Provide(fx.Annotate(
			func(cfg *platform.Config) string {
				if cfg.OutboxQueueURL == "" {
					panic("missing outbox queue URL")
				}
				return cfg.OutboxQueueURL
			},
			fx.ResultTags(`name:"outbox_queue_url"`),
		)),
		fx.Provide(platform.NewOutboxWorker),
		fx.Invoke(func(lc fx.Lifecycle, client *awsSQS.Client, outboxWorker *platform.OutboxWorker, logger *slog.Logger) {
			ctx, cancel := context.WithCancel(context.Background())
			lc.Append(fx.Hook{
				OnStart: func(c context.Context) error {
					logger.Info("Ensuring output SQS queue creation (Outbox Worker)...")
					outDlqName := "wager-events-dlq.fifo"
					outQueueName := "wager-events-outbox.fifo"
					outDlqRes, err := client.CreateQueue(c, &awsSQS.CreateQueueInput{
						QueueName: aws.String(outDlqName),
						Attributes: map[string]string{
							"FifoQueue":                 "true",
							"ContentBasedDeduplication": "true",
						},
					})
					if err != nil {
						logger.Error("Error to create output DLQ in SQS", "error", err)
						return err
					}
					outDlqAttrRes, err := client.GetQueueAttributes(c, &awsSQS.GetQueueAttributesInput{
						QueueUrl: outDlqRes.QueueUrl,
						AttributeNames: []sqsTypes.QueueAttributeName{
							sqsTypes.QueueAttributeNameQueueArn,
						},
					})
					if err != nil {
						logger.Error("Error to query output DLQ ARN", "error", err)
						return err
					}
					outDlqArn := outDlqAttrRes.Attributes["QueueArn"]
					outRedrivePolicy := map[string]string{
						"deadLetterTargetArn": outDlqArn,
						"maxReceiveCount":     "5",
					}
					outPolicyBytes, _ := json.Marshal(outRedrivePolicy)
					_, err = client.CreateQueue(c, &awsSQS.CreateQueueInput{
						QueueName: aws.String(outQueueName),
						Attributes: map[string]string{
							"FifoQueue":                 "true",
							"ContentBasedDeduplication": "true",
							"RedrivePolicy":             string(outPolicyBytes),
						},
					})
					if err != nil {
						logger.Error("Error to create output queue in SQS of outbox", "error", err)
						return err
					}
					logger.Info("Output SQS queues provisioned successfully. Starting Outbox Worker in background...")
					go outboxWorker.Start(ctx)
					return nil
				},
				OnStop: func(c context.Context) error {
					logger.Info("Stopping Outbox Worker...")
					cancel()
					return nil
				},
			})
		}),
	).Run()
}