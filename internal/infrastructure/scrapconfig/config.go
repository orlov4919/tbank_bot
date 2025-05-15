package scrapconfig

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	UpdatesTransport string `env:"UPDATES_TRANSPORT"`
	ScrapperPort     string `env:"SCRAPPER_PORT"`
	GitHubAPIKey     string `env:"GIT_KEY"`
}

func New() (*Config, error) {
	config := &Config{}

	if err := env.Parse(config); err != nil {
		return nil, fmt.Errorf("ошибка при парсинге конфигурации scrapper: %w", err)
	}

	return config, nil
}
