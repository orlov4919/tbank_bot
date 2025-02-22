package telegram

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"linkTraccer/internal/domain/tgbot"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"
)

type Updates = tgbot.Updates

const (
	getUpdates  = "getUpdates"
	sendMessage = "sendMessage"
	jsonType    = "application/json"
)

type TgClient struct {
	basePath string
	scheme   string
	host     string
	client   *http.Client
}

func NewClient(host string) *TgClient {

	return &TgClient{
		scheme: "https",
		host:   host,
		client: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

func (client *TgClient) SetBotToken(token string) {
	client.basePath = "bot" + token
}

func (bot *TgClient) HandleUsersUpdates(offset, limit int) (Updates, error) {

	if limit < 0 {
		return nil, NewErrNegativeLimit(limit)
	}

	var requestURL *url.URL

	q := url.Values{}

	q.Add("offset", strconv.Itoa(offset))
	q.Add("limit", strconv.Itoa(limit))

	requestURL = bot.makeRequestURL(getUpdates, q)
	jsonData, err := RequestToAPI(bot.client, requestURL, http.MethodGet, nil)

	response := &GetUpdateAnswer{}

	if err != nil {
		return nil, fmt.Errorf("при запросе getUpdates произошла ошибка: %w", err)
	}

	if err := json.Unmarshal(jsonData, response); err != nil {
		return nil, fmt.Errorf("при десериализации обновлений произошла ошибка: %w", err)
	}

	return response.Updates, nil
}

func (bot *TgClient) SendMessage(userID int, text string) error {
	sendMessageURL := bot.makeRequestURL(sendMessage, nil)

	data := &SendMessage{
		ID:   userID,
		Text: text,
	}

	jsonData, err := json.Marshal(data)

	if err != nil {
		return fmt.Errorf("при маршалинге сообщения возникла ошибка: %w", err)
	}

	responseData, err := RequestToAPI(bot.client, sendMessageURL, http.MethodPost, bytes.NewBuffer(jsonData))

	if err != nil {
		return fmt.Errorf("при отправке запроса sendMessage произошла ошибка: %w", err)
	}

	serverAnswer := &DefaultServerAnswer{}

	if err := json.Unmarshal(responseData, serverAnswer); err != nil {
		return fmt.Errorf("при декодинге сообщения сервера возникла ошибка: %w", err)
	}

	return nil
}

func RequestToAPI(client *http.Client, url *url.URL, httpMethod string, data io.Reader) ([]byte, error) {

	req, err := http.NewRequest(httpMethod, url.String(), data)

	if err != nil {
		return nil, fmt.Errorf("при создании запроса к botApi возникла ошибка: %w", err)
	}

	if data != nil {
		req.Header.Add("content-type", jsonType)
	}

	r, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("запрос к botAPI закончился ошибкой: %w", err)
	}

	if r.StatusCode != http.StatusOK {
		return nil, errors.New("пришел не 200")
	}

	defer r.Body.Close()

	return io.ReadAll(r.Body)
}

func (bot *TgClient) ChangeSchemeHTTP() {
	bot.scheme = "http"
}

func (bot *TgClient) makeRequestURL(botMethod string, q url.Values) *url.URL {
	return &url.URL{
		Scheme:   bot.scheme,
		Host:     bot.host,
		Path:     path.Join(bot.basePath, botMethod),
		RawQuery: q.Encode(),
	}
}
