package config

import (
	"auth-service/internal/transport/server"
	"auth-service/pkg/db/postgres"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Postgres postgres.Config
	Server   server.ConfigServer
}

func New(path string) (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
