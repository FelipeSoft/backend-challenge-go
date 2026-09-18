package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/FelipeSoft/backend-challenge-go/migrations"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  go run cmd/migrate/main.go up      # Run all pending migrations")
		fmt.Println("  go run cmd/migrate/main.go up 1    # Run only 1 migration forward")
		fmt.Println("  go run cmd/migrate/main.go down 1  # Rollback only 1 migration")
		os.Exit(1)
	}
	command := os.Args[1]
	d, err := iofs.New(migrations.Files, ".")
	if err != nil {
		fmt.Printf("Error loading embedded files: %v\n", err)
		os.Exit(1)
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		panic("missing DATABASE_URL environment variable")
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		fmt.Printf("Error initializing migrate: %v\n", err)
		os.Exit(1)
	}
	switch command {
	case "up":
		if len(os.Args) > 2 {
			steps, err := strconv.Atoi(os.Args[2])
			if err != nil {
				fmt.Println("The number of steps must be a valid integer.")
				os.Exit(1)
			}
			err = m.Steps(steps)
			if err != nil && !errors.Is(err, migrate.ErrNoChange) {
				fmt.Printf("Error executing %d step(s) forward: %v\n", steps, err)
				os.Exit(1)
			}
			fmt.Printf("Success: %d migration(s) applied!\n", steps)
			return
		}
		err = m.Up()
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			fmt.Printf("Error running all migrations: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Success: All pending migrations were applied!")
	case "down":
		steps := -1
		if len(os.Args) > 2 {
			s, err := strconv.Atoi(os.Args[2])
			if err != nil {
				fmt.Println("The number of steps must be a valid integer.")
				os.Exit(1)
			}
			steps = -s
		}
		err = m.Steps(steps)
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			fmt.Printf("Error executing rollback: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Success: Rollback executed!")
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}