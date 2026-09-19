package platform

import (
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/fx"
)

type Config struct {
	LocalStackAuthToken string
	Port                string
	DatabaseURL         string
	SQSQueueURL         string
	OutboxQueueURL      string
	AwsAccessKeyId      string
	AwsSecretAccessKey  string
	AwsRegion           string
	SQSBaseEndpoint     string
}

var ConfigModule = fx.Options(
	fx.Provide(Load),
)

func Load() *Config {
	_ = godotenv.Load()
	cfg := &Config{
		LocalStackAuthToken: os.Getenv("LOCALSTACK_AUTH_TOKEN"),
		Port:                os.Getenv("PORT"),
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		SQSQueueURL:         os.Getenv("SQS_QUEUE_URL"),
		OutboxQueueURL:      os.Getenv("OUTBOX_QUEUE_URL"),
		AwsRegion:           os.Getenv("AWS_REGION"),
		AwsAccessKeyId:      os.Getenv("AWS_ACCESS_KEY_ID"),
		AwsSecretAccessKey:  os.Getenv("AWS_SECRET_ACCESS_KEY"),
		SQSBaseEndpoint:     os.Getenv("SQS_BASE_ENDPOINT"),
	}
	return cfg
}
