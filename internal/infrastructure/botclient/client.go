package botclient

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	jsonData, err := json.Marshal(update)

	if err != nil {
		return fmt.Errorf("ошибка маршалинга обновлений: %w", err)
	}

	tgBotURL := &url.URL{
		Scheme: "http",
		Host:   bot.host,
		Path:   bot.basePath,
	}

	req, err := http.NewRequest(http.MethodPost, tgBotURL.String(), bytes.NewBuffer(jsonData))

	if err != nil {
		return fmt.Errorf("ошибка при составлении запроса для отправки обновлений: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := bot.client.Do(req)

	if err != nil {
		return fmt.Errorf("ошибка при отправке обновлений боту: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return NewErrBadAnswerFromServer(resp.StatusCode)
	}

	return nil
}
