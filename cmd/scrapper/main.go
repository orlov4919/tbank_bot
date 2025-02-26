package main

import (
	"encoding/json"
	"fmt"
	"linkTraccer/internal/domain/scrapper"
)

func main() {

	//stackClient := stackoverflow.NewClient("api.stackexchange.com", &http.Client{Timeout: time.Second * 10})
	//botClient := botClient.New(":8080", &http.Client{Timeout: time.Second * 10})
	//userRepo := file.NewFileStorage()
	//
	//userRepo.TrackLink(771592675, "https://stackoverflow.com/questions/79463686/split-string-into-columns-dynamically-create-columns-based-on-length-of-string")
	//userRepo.TrackLink(771592675, "https://stackoverflow.com/questions/79463690/renaming-a-single-column-with-unknown-name-in-polars-rust")
	//userRepo.TrackLink(771592675, "https://stackoverflow.com/questions/79463691/how-to-find-non-declared-css-within-visual-studio")
	//scrapper := scrapperService.New(userRepo, botClient, stackClient)
	//
	//scrapper.CheckLinkUpdates()

	//gitClient := gitHub.NewClient("api.github.com", "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", &http.Client{Timeout: time.Minute})
	//
	//fmt.Println(gitClient.LinkState("https://github.com/orlov4919/test"))

	resp := scrapper.LinkResponse{
		ID:      5,
		URL:     "hello",
		Tags:    []string{"a", "b"},
		Filters: []string{"a", "b"},
	}

	list := &scrapper.ListLinksResponse{
		Links: []scrapper.LinkResponse{resp},
		Size:  1,
	}

	data, _ := json.Marshal(list)

	fmt.Println(string(data))
}
