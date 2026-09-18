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
					logger.Info("Garantindo a criação das filas SQS FIFO e DLQ...")
					dlqName := "wager-transactions-dlq.fifo"
					mainQueueName := "wager-transactions.fifo"
					dlqRes, err := client.CreateQueue(c, &awsSQS.CreateQueueInput{
						QueueName: aws.String(dlqName),
						Attributes: map[string]string{
							"FifoQueue":                 "true",
							"ContentBasedDeduplication": "true",
						},
					})
					if err != nil {
						logger.Error("Erro ao criar DLQ SQS", "error", err)
						return err
					}
					attrRes, err := client.GetQueueAttributes(c, &awsSQS.GetQueueAttributesInput{
						QueueUrl: dlqRes.QueueUrl,
						AttributeNames: []sqsTypes.QueueAttributeName{
							sqsTypes.QueueAttributeNameQueueArn,
						},
					})
					if err != nil {
						logger.Error("Erro ao buscar ARN da DLQ", "error", err)
						return err
					}
					dlqArn := attrRes.Attributes["QueueArn"]
					redrivePolicy := map[string]string{
						"deadLetterTargetArn": dlqArn,
						"maxReceiveCount":     "5",
					}
					policyBytes, _ := json.Marshal(redrivePolicy)
					_, err = client.CreateQueue(c, &awsSQS.CreateQueueInput{
						QueueName: aws.String(mainQueueName),
						Attributes: map[string]string{
							"FifoQueue":                 "true",
							"ContentBasedDeduplication": "true",
							"RedrivePolicy":             string(policyBytes),
						},
					})
					if err != nil {
						logger.Error("Erro ao criar fila principal SQS", "error", err)
						return err
					}

					logger.Info("Filas SQS provisionadas com sucesso. Iniciando o consumer em background...")
					go func() {
						if err := consumer.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
							logger.Error("Consumer SQS encerrou com erro crítico", "error", err)
						}
					}()
					return nil
				},
				OnStop: func(c context.Context) error {
					logger.Info("Sinal SIGTERM recebido. Parando o consumer SQS graciosamente...")
					cancel()
					stopCtx, stopCancel := context.WithTimeout(c, 10*time.Second)
					defer stopCancel()
					if err := consumer.Stop(stopCtx); err != nil {
						logger.Error("Erro ao encerrar o consumer SQS", "error", err)
						return err
					}
					logger.Info("Consumer SQS finalizado com sucesso.")
					return nil
				},
			})
		}),
	).Run()
}