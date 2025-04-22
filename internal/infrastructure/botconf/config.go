package botconf

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	BotToken         string `env:"BOT_TOKEN"`
	ScrapperPort     string `env:"SCRAPPER_PORT"`
	ScrapperHost     string `env:"SCRAPPER_HOST"`
	BotPort          string `env:"BOT_PORT"`
	BotBatch         int    `env:"BOT_BATCH"`
	UpdatesTransport string `env:"UPDATES_TRANSPORT"`
}

func New() (*Config, error) {
	config := &Config{}

	if err := env.Parse(config); err != nil {
		return nil, fmt.Errorf("ошибка при конфигурации: %w", err)
	}

	return config, nil
}
