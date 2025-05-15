package producer

import (
	"fmt"
	"github.com/caarlos0/env/v11"
)

// в качестве адресов брокеров, указывается строка, где адреса разделены запятой без пробелов

type Config struct {
	Topic   string `env:"UPDATE_TOPIC"`
	Brokers string `env:"BROKERS_ADDR"`
	Bath    int    `env:"KAFKA_BATCH_SIZE"`
}

func NewConfig() (*Config, error) {
	config := &Config{}

	if err := env.Parse(config); err != nil {
		return nil, fmt.Errorf("ошибка при парсинге конфига кафка продюсера: %w", err)
	}

	return config, nil
}
