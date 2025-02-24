package main

import (
	"linkTraccer/internal/application/botService"
	"linkTraccer/internal/infrastructure/telegram"
)

const (
	pathBotConfig    = "../../configs/bot/config.toml"
	host             = "api.telegram.org"
	errChanSize      = 10000
	updateServerAddr = ":8080"
)

func main() {

	tgClient := telegram.NewClient("api.telegram.org")
	tgClient.SetBotToken("7723330730:AAEKTKDWzoEwWK4DrLqWC4q-7LsqwXfOMUU")

	tgBot := botService.New(tgClient, 5)
	tgBot.Start()

	//userRepo := file.NewFileStorage()
}
