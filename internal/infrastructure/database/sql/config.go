package sql

import (
	"fmt"
	"github.com/caarlos0/env/v11"
)

type DBConfig struct {
	DBUser     string `env:"DB_USER"`
	DBPass     string `env:"DB_PASS"`
	DBName     string `env:"DB_NAME"`
	DBHost     string `env:"DB_HOST"`
	DBPort     string `env:"DB_PORT"`
	BatchSize  int    `env:"BATCH_SIZE"`
	AccessType string `env:"ACCESS_TYPE"`
}

func NewConfig() (*DBConfig, error) {
	config := &DBConfig{}

	if err := env.Parse(config); err != nil {
		return nil, fmt.Errorf("во время декодинга конфига базы данных произошла ошибка : %w", err)
	}

	return config, nil
}

func (d *DBConfig) ToDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s%s/%s?sslmode=disable", d.DBUser, d.DBPass, d.DBHost, d.DBPort, d.DBName)
}
