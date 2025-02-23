package botClient

import (
	"encoding/json"
	"linkTraccer/internal/domain/dto"
	"net/http"
)

const (
	updatesPath = "/updates"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}


type BotClient struct {
	host     string
	basePath string
	client HTTPClient
}

func New(host string, client HTTPClient) *BotClient {
	return &BotClient{
		host:     host,
		basePath: updatesPath,
		client : client,
	}
}

func (bot *BotClient) SendLinkUpdates(update dto.LinkUpdate) error {

	tgBotUrl :=

	jsonData := json.Marshal(update, )

	req, _  := http.NewRequest(http.MethodPost,tgBotUrl, )

}
