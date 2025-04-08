package scrapconfig

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	BotHost      string `env:"BOT_HOST"`
	BotPort      string `env:"BOT_PORT"`
	ScrapperPort string `env:"SCRAPPER_PORT"`
	GitHubAPIKey string `env:"GIT_KEY"`
}

func New() (*Config, error) {
	config := &Config{}

	if err := env.Parse(config); err != nil {
		return nil, fmt.Errorf("ошибка при конфигурации scrapper: %w", err)
	}

	return config, nil
}
