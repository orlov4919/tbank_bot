package main

import (
	"linkTraccer/internal/application/botService"
	"linkTraccer/internal/infrastructure/botconfig"
	"linkTraccer/internal/infrastructure/bothandler"
	"linkTraccer/internal/infrastructure/database/file/contextstorage"
	"linkTraccer/internal/infrastructure/scrapperclient"
	"linkTraccer/internal/infrastructure/telegram"
	"log"
	"net/http"
	"time"
)

const (
	telegramBotAPI = "api.telegram.org"
)

func main() {
	config, err := botconfig.New()

	if err != nil {
		log.Fatal(err)
	}

	tgClient := telegram.NewClient(&http.Client{Timeout: time.Minute}, config.Token, telegramBotAPI)
	ctxStore := contextstorage.New()
	scrapClient := scrapperclient.New(&http.Client{Timeout: time.Minute}, config.ScrapperServerUrl)
	tgBot := botService.New(tgClient, scrapClient, ctxStore, 5)

	go tgBot.Start()

	mux := http.NewServeMux()

	mux.HandleFunc("/updates", bothandler.New(tgClient).HandleLinkUpdates)

	http.ListenAndServe(config.BotServerPort, mux)
}
