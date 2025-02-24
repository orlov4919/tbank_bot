package botClient

import (
	"bytes"
	"encoding/json"
	"errors"
	"linkTraccer/internal/domain/dto"
	"net/http"
	"net/url"
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
	client   HTTPClient
}

func New(host string, client HTTPClient) *BotClient {
	return &BotClient{
		host:     host,
		basePath: updatesPath,
		client:   client,
	}
}

func (bot *BotClient) SendLinkUpdates(update *dto.LinkUpdate) error {

	tgBotUrl := &url.URL{
		Scheme: "http",
		Host:   bot.host,
		Path:   bot.basePath,
	}

	jsonData, err := json.Marshal(update)

	if err != nil {
		return errors.New("Ошибка маршалинге запроса боту")
	}

	req, err := http.NewRequest(http.MethodPost, tgBotUrl.String(), bytes.NewBuffer(jsonData))

	if err != nil {
		return errors.New("Ошибка создании запроса боту")
	}

	resp, err := bot.client.Do(req)

	if err != nil {
		return errors.New("Ошибка при отправлении запроса боту")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("От сервера бота пришла не 200")
	}

	return nil
}
