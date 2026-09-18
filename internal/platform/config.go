package platform

import (
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/fx"
)

type Config struct {
	AppEnv      string
	Port        string
	DatabaseURL string
	SQSEndpoint string
}

var ConfigModule = fx.Options(
	fx.Provide(Load),
)

func Load() *Config {
	_ = godotenv.Load()
	cfg := &Config{
		AppEnv:      os.Getenv("APP_ENV"),
		Port:        os.Getenv("PORT"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		SQSEndpoint: os.Getenv("SQS_ENDPOINT"),
	}
	return cfg
}