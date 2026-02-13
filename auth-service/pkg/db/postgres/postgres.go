package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host           string `yaml:"POSTGRES_HOST" env:"POSTGRES_HOST"`
	Port           int    `yaml:"POSTGRES_PORT" env:"POSTGRES_PORT"`
	User           string `yaml:"POSTGRES_USER" env:"POSTGRES_USER"`
	Password       string `yaml:"POSTGRES_PASSWORD" env:"POSTGRES_PASSWORD"`
	DataBase       string `yaml:"POSTGRES_DB" env:"POSTGRES_DB"`
	MinConnections int    `yaml:"POSTGRES_MIN_CONNECTIONS" env:"POSTGRES_MIN_CONNECTIONS"`
	MaxConnections int    `yaml:"POSTGRES_MAX_CONNECTIONS" env:"POSTGRES_MAX_CONNECTIONS"`
}

type DataBase struct {
	Pool *pgxpool.Pool
}

// postgres://username:password@localhost:5432/database_name
func New(ctx context.Context, config Config) (DataBase, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&pool_max_conns=%d&pool_min_conns=%d",
		config.User, config.Password, config.Host, config.Port, config.DataBase, config.MaxConnections, config.MinConnections)
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return DataBase{}, fmt.Errorf("db.New: %w", err)
	}
	err = pool.Ping(ctx)
	if err != nil {
		return DataBase{}, fmt.Errorf("db.Ping: %w", err)
	}
	return DataBase{
		Pool: pool,
	}, nil
}
func (db *DataBase) Close() {
	db.Pool.Close()
}
