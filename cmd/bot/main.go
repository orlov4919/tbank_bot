package main

import (
	"linkTraccer/internal/application/botService"
	"linkTraccer/internal/infrastructure/botHandler"
	"linkTraccer/internal/infrastructure/database/file/contextStorage"
	"linkTraccer/internal/infrastructure/scrapperClient"
	"linkTraccer/internal/infrastructure/telegram"
	"net/http"
	"time"
)

const (
	pathBotConfig    = "../../configs/bot/config.toml"
	host             = "api.telegram.org"
	errChanSize      = 10000
	updateServerAddr = ":8080"
)

func main() {

	tgClient := telegram.NewClient(&http.Client{Timeout: time.Minute}, "api.telegram.org")
	tgClient.SetBotToken("7723330730:AAEKTKDWzoEwWK4DrLqWC4q-7LsqwXfOMUU") // убрать эту ерунду

	ctxStore := contextStorage.New()

	scrapClient := scrapperClient.New(&http.Client{Timeout: time.Minute}, "localhost:9090")

	tgBot := botService.New(tgClient, scrapClient, ctxStore, 5)

	go tgBot.Start()

	mux := http.NewServeMux()

	mux.HandleFunc("/updates", botHandler.New(tgClient).HandleLinkUpdates)

	http.ListenAndServe(":8080", mux)
}
