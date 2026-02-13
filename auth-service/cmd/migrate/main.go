package main

import (
	"auth-service/internal/config"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/joho/godotenv"
)

func main() {
	var migrationsPath string
	var command string

	flag.StringVar(&migrationsPath, "path", "./migrations", "path to migrations directory")
	flag.StringVar(&command, "command", "up", "migration command: up, down, force, version")
	flag.Parse()

	envPath := os.Getenv("ENV_PATH")
	if envPath == "" {
		envPath = "./config/.env"
	}

	if err := godotenv.Load(envPath); err != nil {
		fmt.Printf("warning: could not load .env file from %s: %v\n", envPath, err)
	}
	fmt.Println("=== Environment variables ===")
	fmt.Printf("POSTGRES_USER: %s\n", os.Getenv("POSTGRES_USER"))
	fmt.Printf("POSTGRES_PASSWORD: %s\n", os.Getenv("POSTGRES_PASSWORD"))
	fmt.Printf("POSTGRES_HOST: %s\n", os.Getenv("POSTGRES_HOST"))
	fmt.Printf("POSTGRES_PORT: %s\n", os.Getenv("POSTGRES_PORT"))
	fmt.Printf("POSTGRES_DB: %s\n", os.Getenv("POSTGRES_DB"))
	fmt.Println("=============================")
	cfg, err := config.New(envPath)
	if err != nil {
		fmt.Printf("error parsing config: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Config: %+v\n", cfg)
	// postgres://username:password@localhost:5432/database_name
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.DataBase,
	)

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		dsn,
	)
	if err != nil {
		fmt.Printf("failed to create migrate instance: %v\n", err)
		os.Exit(1)
	}

	defer m.Close()

	switch command {
	case "up":
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			fmt.Printf("failed to apply migrations: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Migrations applied successfully!")

	case "down":
		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			fmt.Printf("failed to rollback migrations: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Migrations rolled back successfully!")

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			fmt.Printf("failed to get version: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Current version: %d, Dirty: %v\n", version, dirty)

	default:
		fmt.Printf("unknown command: %s\n", command)
		fmt.Println("Available commands: up, down, version")
		os.Exit(1)
	}
}
