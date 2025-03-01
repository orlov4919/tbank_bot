package scrapperconfig

import "github.com/BurntSushi/toml"

type Config struct {
	TgBotServerURL     string `toml:"tgbot_server_url"`
	ScrapperServerPort string `toml:"scrapper_server_port"`
	GitHubToken        string `toml:"github_token"`
}

const (
	pathToConfig = "../../configs/scrapper/config.toml"
)

func New() (*Config, error) {
	config := &Config{}
	_, err := toml.DecodeFile(pathToConfig, config)

	return config, err
}
