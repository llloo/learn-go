package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	ServerPort  string `envconfig:"APP_SERVER_PORT" default:"8080"`
	DatabaseURL string `envconfig:"APP_DATABASE_URL" default:"postgresql://postgres:psql@localhost:5432/myapp?sslmode=disable"`
	JWTSecret   string `envconfig:"APP_JWT_SECRET" default:"your-secret-key"`
}

func NewConfig() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
