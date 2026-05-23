package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	ServerPort  string `envconfig:"APP_SERVER_PORT" default:"8080"`
	DatabaseURL string `envconfig:"APP_DATABASE_URL" default:"postgres://psql:psql@localhost:5432/psql?sslmode=disable"`
}

func NewConfig() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
