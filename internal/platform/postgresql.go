package platform

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

var DatabaseModule = fx.Options(
	fx.Provide(NewPostgresPool),
)

func NewPostgresPool(lc fx.Lifecycle) *pgxpool.Pool {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		panic("the DATABASE_URL environment variable is required")
	}
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		panic(fmt.Sprintf("failed to parse DATABASE_URL: %v", err))
	}
	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 30 * time.Minute
	poolConfig.MaxConnIdleTime = 5 * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("falha ao criar pool de conexões do postgres: %v", err)
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return pool.Ping(ctx)
		},
		OnStop: func(ctx context.Context) error {
			pool.Close()
			return nil
		},
	})
	return pool
}