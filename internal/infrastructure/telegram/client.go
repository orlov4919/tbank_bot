package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"linkTraccer/internal/domain/tgbot"
	"net/http"
	"net/url"
	"path"
	"strconv"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

const (
	setMyCommands = "setMyCommands"
	getUpdates    = "getUpdates"
	sendMessage   = "sendMessage"
	jsonType      = "application/json"
)

type TgClient struct {
	basePath string
	scheme   string
	host     string
	client   HTTPClient
}

func NewClient(client HTTPClient, token, host string) *TgClient {
	return &TgClient{
		scheme:   "https",
		host:     host,
		client:   client,
		basePath: "bot" + token,
	}
}

func (bot *TgClient) HandleUsersUpdates(offset, limit int) (tgbot.Updates, error) {
	if limit < 0 {
		return nil, NewErrNegativeLimit(limit)
	}

	var requestURL *url.URL

	q := url.Values{}

	q.Add("offset", strconv.Itoa(offset))
	q.Add("limit", strconv.Itoa(limit))

	requestURL = bot.makeRequestURL(getUpdates, q)
	jsonData, err := RequestToAPI(bot.client, requestURL, http.MethodGet, nil)

	if err != nil {
		return nil, fmt.Errorf("при запросе getUpdates к tg API произошла ошибка: %w", err)
	}

	response := &GetUpdateAnswer{}

	if err := json.Unmarshal(jsonData, response); err != nil {
		return nil, fmt.Errorf("при десериализации обновлений от сервера телеграм произошла ошибка: %w", err)
	}

	return response.Updates, nil
}

func (bot *TgClient) SendMessage(userID int64, text string) error {
	sendMessageURL := bot.makeRequestURL(sendMessage, nil)

	data := &SendMessage{
		ID:   userID,
		Text: text,
	}

	jsonData, err := json.Marshal(data)

	if err != nil {
		return fmt.Errorf("при маршалинге сообщения, для отправки на сервер телеграмм возникла ошибка: %w", err)
	}

	_, err = RequestToAPI(bot.client, sendMessageURL, http.MethodPost, bytes.NewBuffer(jsonData))

	if err != nil {
		return fmt.Errorf("при отправке сообщения на сервер телеграмм произошла ошибка: %w", err)
	}

	return nil
}

func (bot *TgClient) SetBotCommands(data *tgbot.SetCommands) error {
	setCommandsURL := bot.makeRequestURL(setMyCommands, nil)
	jsonData, err := json.Marshal(data)

	if err != nil {
		return fmt.Errorf("при маршалинге команд поддерживаемых ботом возникла ошибка: %w", err)
	}

	_, err = RequestToAPI(bot.client, setCommandsURL, http.MethodPost, bytes.NewBuffer(jsonData))

	if err != nil {
		return fmt.Errorf("запрос на создание меню с командами закончился ошибкой: %w", err)
	}

	return nil
}

func RequestToAPI(client HTTPClient, url *url.URL, httpMethod string, data io.Reader) ([]byte, error) {
	req, err := http.NewRequest(httpMethod, url.String(), data)

	if err != nil {
		return nil, fmt.Errorf("при создании запроса к botApi возникла ошибка: %w", err)
	}

	if data != nil {
		req.Header.Add("Content-Type", jsonType)
	}

	r, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("запрос к botAPI закончился ошибкой: %w", err)
	}

	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, NewErrBotAPI(r.StatusCode)
	}

	return io.ReadAll(r.Body)
}

func (bot *TgClient) makeRequestURL(botMethod string, q url.Values) *url.URL {
	return &url.URL{
		Scheme:   bot.scheme,
		Host:     bot.host,
		Path:     path.Join(bot.basePath, botMethod),
		RawQuery: q.Encode(),
	}
}
