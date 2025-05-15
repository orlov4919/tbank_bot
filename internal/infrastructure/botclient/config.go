package botclient

import (
	"fmt"
	"github.com/caarlos0/env/v11"
)

type Config struct {
	BotHost string `env:"BOT_HOST"`
	BotPort string `env:"BOT_PORT"`
}

func NewConfig() (*Config, error) {
	config := &Config{}

	if err := env.Parse(config); err != nil {
		return nil, fmt.Errorf("ошибка при парсинге конфига кафка продюсера: %w", err)
	}

	return config, nil
}
