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
	}
	return cfg
}
