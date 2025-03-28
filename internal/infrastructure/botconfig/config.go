package botconfig

import (
	"fmt"
	"github.com/caarlos0/env/v11"
)

type Config struct {
	Token        string `env:"BOT_TOKEN"`
	ScrapperPort string `env:"SCRAPPER_PORT"`
	ScrapperHost string `env:"SCRAPPER_HOST"`
	BotPort      string `env:"BOT_PORT"`
}

func New() (*Config, error) {
	config := &Config{}

	if err := env.Parse(config); err != nil {
		return nil, fmt.Errorf("ошибка при конфигурации: %w", err)
	}

	return config, nil
}
