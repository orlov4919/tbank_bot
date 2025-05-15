package redis

import (
	"fmt"
	"github.com/caarlos0/env/v11"
)

type Config struct {
	RedisAddr string `env:"REDIS_ADDR"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("ошибка при парсинге конфига: %w", err)
	}

	return cfg, nil
}
