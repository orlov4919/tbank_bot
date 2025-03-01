package botconfig

import "github.com/BurntSushi/toml"

type Config struct {
	Token             string `toml:"token"`
	ScrapperServerUrl string `toml:"scrapper_server_url"`
	BotServerPort     string `toml:"bot_server_port"`
}

const (
	pathToConfig = "../../configs/bot/config.toml"
)

func New() (*Config, error) {
	config := &Config{}
	_, err := toml.DecodeFile(pathToConfig, config)

	return config, err
}
