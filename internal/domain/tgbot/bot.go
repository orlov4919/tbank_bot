package tgbot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"
)

const (
	getUpdates  = "getUpdates"
	sendMessage = "sendMessage"
	jsonType    = "application/json"
)

type TgBot struct {
	token    string
	offset   int
	limit    int
	host     string
	basePath string
	scheme   string
	client   *http.Client
}

type Config struct {
	Token string `toml:"token"`
}

func New(config *Config, host string) *TgBot {
	return &TgBot{
		token:    config.Token,
		offset:   0,
		host:     host,
		basePath: "bot" + config.Token,
		limit:    100,
		client: &http.Client{
			Timeout: time.Second * 10,
		},
		scheme: "https",
	}
}

func NewConfig() *Config {
	return &Config{}
}

func (bot *TgBot) HandleUsersUpdates() (Updates, error) {
	var requestURL *url.URL

	q := url.Values{}

	q.Add("offset", strconv.Itoa(bot.offset))
	q.Add("limit", strconv.Itoa(bot.limit))

	requestURL = bot.makeRequestURL(getUpdates, q)
	jsonData, err := RequestToBotAPI(bot.client, http.MethodGet, nil, requestURL)

	response := &GetUpdateAnswer{}

	if err != nil {
		return nil, fmt.Errorf("при запросе getUpdates произошла ошибка: %w", err)
	}

	if err := json.Unmarshal(jsonData, response); err != nil {
		return nil, fmt.Errorf("при десериализации обновлений произошла ошибка: %w", err)
	}

	fmt.Println(response.Updates)

	return response.Updates, nil
}

func RequestToBotAPI(client *http.Client, httpMethod string, data io.Reader, url *url.URL) ([]byte, error) {
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

func (bot *TgBot) ChangeOffset(newOffset int) {
	bot.offset = newOffset
}

func (bot *TgBot) SendMessage(userID int, text string) error {
	sendMessageURL := bot.makeRequestURL(sendMessage, nil)

	data := &SendMessage{
		ID:   userID,
		Text: text,
	}

	jsonData, err := json.Marshal(data)

	if err != nil {
		return fmt.Errorf("при маршалинге сообщения возникла ошибка: %w", err)
	}

	responseData, err := RequestToBotAPI(bot.client, http.MethodPost, bytes.NewBuffer(jsonData), sendMessageURL)

	if err != nil {
		return fmt.Errorf("при отправке запроса sendMessage произошла ошибка: %w", err)
	}

	serverAnswer := &DefaultServerAnswer{}

	if err := json.Unmarshal(responseData, serverAnswer); err != nil {
		return fmt.Errorf("при декодинге сообщения сервера возникла ошибка: %w", err)
	}

	return nil
}

func (bot *TgBot) ChangeScheme(scheme string) {
	bot.scheme = scheme
}

func (bot *TgBot) makeRequestURL(botMethod string, q url.Values) *url.URL {
	return &url.URL{
		Scheme:   bot.scheme,
		Host:     bot.host,
		Path:     path.Join(bot.basePath, botMethod),
		RawQuery: q.Encode(),
	}
}
