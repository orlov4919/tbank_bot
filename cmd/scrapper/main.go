package main

import (
	"github.com/gorilla/mux"
	"linkTraccer/internal/application/scrapperService"
	"linkTraccer/internal/infrastructure/botClient"
	"linkTraccer/internal/infrastructure/database/file/userStorage"
	"linkTraccer/internal/infrastructure/scrapperHandlers"
	"linkTraccer/internal/infrastructure/siteClients/gitHub"
	"linkTraccer/internal/infrastructure/siteClients/stackoverflow"
	"net/http"
	"time"
)

const (
	stackOverflowAPI = "api.stackexchange.com"
	gitHubApi        = "api.github.com"
)

func main() {

	stackClient := stackoverflow.NewClient(stackOverflowAPI, &http.Client{Timeout: time.Second * 10})
	botClient := botClient.New("localhost:8080", &http.Client{Timeout: time.Second * 10})
	userRepo := userStorage.NewFileStorage()
	gitClient := gitHub.NewClient("api.github.com",
		"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",
		&http.Client{Timeout: time.Minute})

	scrapper := scrapperService.New(userRepo, botClient, stackClient, gitClient)

	go scrapper.CheckLinkUpdates()

	mux := mux.NewRouter()
	linksHandler := scrapperHandlers.NewLinkHandler(userRepo, stackClient, gitClient)
	chatHandler := scrapperHandlers.NewChatHandler(userRepo)

	mux.HandleFunc("/tg-chat/{id}", chatHandler.HandleChatChanges)
	mux.HandleFunc("/links", linksHandler.HandleLinksChanges)

	http.ListenAndServe(":9090", mux)

}
