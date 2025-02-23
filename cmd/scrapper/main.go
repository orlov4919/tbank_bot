package main

import (
	"fmt"
	"linkTraccer/internal/infrastructure/siteClients/stackoverflow"
	"net/http"
	"time"
)

func main() {

	stackClient := stackoverflow.NewClient("api.stackexchange.com", &http.Client{Timeout: time.Second * 10})

	fmt.Println(stackClient.LinkState("https://stackoverflow.com/questions/259123/how-do-i-safely-populate-with-data-and-refresh-a-datagridview-in-a-multi-threa/768143#768143"))

}
