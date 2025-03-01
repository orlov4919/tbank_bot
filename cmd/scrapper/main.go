package main

import (
	"github.com/go-co-op/gocron"
	"linkTraccer/internal/application/scrapperService"
	"linkTraccer/internal/infrastructure/botclient"
	"linkTraccer/internal/infrastructure/database/file/userstorage"
	"linkTraccer/internal/infrastructure/scrapperconfig"
	"linkTraccer/internal/infrastructure/scrapperhandlers"
	"linkTraccer/internal/infrastructure/siteclients/github"
	"linkTraccer/internal/infrastructure/siteclients/stackoverflow"
	"log"
	"net/http"
	"time"
)

const (
	stackOverflowAPI = "api.stackexchange.com"
	gitHubApi        = "api.github.com"
)

func main() {
	config, err := scrapperconfig.New()

	if err != nil {
		log.Println(err)
		log.Fatal("Ошибка при ините конфига")
	}

	stackClient := stackoverflow.NewClient(stackOverflowAPI, &http.Client{Timeout: time.Second * 10})
	tgBotClient := botclient.New(config.TgBotServerURL, &http.Client{Timeout: time.Second * 10})
	userRepo := userstorage.NewFileStorage()
	gitClient := github.NewClient(gitHubApi, config.GitHubToken, &http.Client{Timeout: time.Minute})

	scrapper := scrapperService.New(userRepo, tgBotClient, stackClient, gitClient)

	s := gocron.NewScheduler(time.UTC)

	_, err = s.Every(time.Minute).Do(scrapper.CheckLinkUpdates)

	if err != nil {
		log.Println("Ошибкаааа")
	}

	s.StartAsync()

	mux := http.NewServeMux()
	linksHandler := scrapperhandlers.NewLinkHandler(userRepo, stackClient, gitClient)
	chatHandler := scrapperhandlers.NewChatHandler(userRepo)

	mux.HandleFunc("/tg-chat/", chatHandler.HandleChatChanges)
	mux.HandleFunc("/links", linksHandler.HandleLinksChanges)

	http.ListenAndServe(config.ScrapperServerPort, mux)
}
