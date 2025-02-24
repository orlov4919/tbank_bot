package main

import (
	"linkTraccer/internal/application/scrapperService"
	"linkTraccer/internal/infrastructure/botClient"
	"linkTraccer/internal/infrastructure/database/file"
	"linkTraccer/internal/infrastructure/siteClients/stackoverflow"
	"net/http"
	"time"
)

func main() {

	stackClient := stackoverflow.NewClient("api.stackexchange.com", &http.Client{Timeout: time.Second * 10})
	botClient := botClient.New(":8080", &http.Client{Timeout: time.Second * 10})
	userRepo := file.NewFileStorage()

	userRepo.TrackLink(771592675, "https://stackoverflow.com/questions/79463686/split-string-into-columns-dynamically-create-columns-based-on-length-of-string")
	userRepo.TrackLink(771592675, "https://stackoverflow.com/questions/79463690/renaming-a-single-column-with-unknown-name-in-polars-rust")
	userRepo.TrackLink(771592675, "https://stackoverflow.com/questions/79463691/how-to-find-non-declared-css-within-visual-studio")
	scrapper := scrapperService.New(userRepo, botClient, stackClient)

	scrapper.CheckLinkUpdates()
}
