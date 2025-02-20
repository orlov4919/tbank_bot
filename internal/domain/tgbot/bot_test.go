package tgbot_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"linkTraccer/internal/domain/tgbot"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
)

const (
	configPath         = "../../../configs/bot/config.toml"
	getUpdates         = "/getUpdates"
	randomStr          = "sdfklsjc"
	returnIncorectJSON = "верни неправильный json"
	content            = "content-type"
	jsonType           = "application/json"
)

func TestRequestToBotAPI(t *testing.T) {
	type TestCase struct {
		method  string
		path    string
		correct bool
		client  *http.Client
		data    io.Reader
	}

	testCases := []TestCase{
		{
			method:  http.MethodGet,
			path:    "/getUpdates",
			client:  http.DefaultClient,
			data:    nil,
			correct: true,
		},
		{
			method:  http.MethodGet,
			path:    "/Updates",
			client:  http.DefaultClient,
			data:    nil,
			correct: false,
		},

		{
			method:  "//",
			path:    "/getUpdates",
			client:  http.DefaultClient,
			data:    nil,
			correct: false,
		},

		{
			method:  "Posts",
			path:    "/getUpdates",
			client:  http.DefaultClient,
			data:    nil,
			correct: false,
		},

		{
			method:  http.MethodPost,
			path:    "/sendMessage",
			client:  http.DefaultClient,
			data:    &bytes.Buffer{},
			correct: true,
		},

		{
			method:  http.MethodPost,
			path:    "/sendMessage",
			client:  &http.Client{Timeout: time.Nanosecond},
			data:    nil,
			correct: false,
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/getUpdates", UpdatesHandler)
	mux.HandleFunc("/sendMessage", MessageHandler)

	server := httptest.NewServer(mux)

	defer server.Close()

	for _, test := range testCases {
		parsedURL, _ := url.Parse(server.URL + test.path)
		_, err := tgbot.RequestToBotAPI(test.client, test.method, test.data, parsedURL)

		if test.correct {
			assert.NoError(t, err)
		} else {
			fmt.Println(err)
			assert.Error(t, err)
		}
	}
}

func TestTgBot_HandleUsersUpdates(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	botConfig := tgbot.NewConfig()

	defer server.Close()

	if _, err := toml.DecodeFile(configPath, botConfig); err != nil {
		log.Println(fmt.Errorf("при парсинге конфига возникла ошибка: %w", err))
	}

	mux.HandleFunc("/bot"+botConfig.Token+"/getUpdates/", getUpdatesHandler)

	serverURL, _ := url.Parse(server.URL)
	serverHost := serverURL.Host

	type TestCase struct {
		bot     *tgbot.TgBot
		correct bool
	}

	testCases := []TestCase{
		{bot: tgbot.New(botConfig, serverHost),
			correct: true,
		},
		{
			bot:     tgbot.New(botConfig, serverHost+randomStr),
			correct: false,
		},
	}

	for _, test := range testCases {
		test.bot.ChangeScheme("http")

		_, err := test.bot.HandleUsersUpdates()

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestTgBot_SendMessage(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	botConfig := tgbot.NewConfig()

	defer server.Close()

	if _, err := toml.DecodeFile(configPath, botConfig); err != nil {
		log.Println(fmt.Errorf("при парсинге конфига возникла ошибка: %w", err))
	}

	mux.HandleFunc("/bot"+botConfig.Token+"/sendMessage", sendMessageHandler)

	serverURL, _ := url.Parse(server.URL)
	serverHost := serverURL.Host

	type TestCase struct {
		bot     *tgbot.TgBot
		id      int
		text    string
		correct bool
	}

	testCases := []TestCase{

		{bot: tgbot.New(botConfig, serverHost),
			id:      15,
			text:    "Привет мир",
			correct: true,
		},
		{
			bot:     tgbot.New(botConfig, serverHost),
			id:      -225,
			text:    "Привет мир",
			correct: false,
		},
		{
			bot:     tgbot.New(botConfig, serverHost),
			id:      228,
			text:    returnIncorectJSON,
			correct: false,
		},
	}

	for _, test := range testCases {
		test.bot.ChangeScheme("http")

		err := test.bot.SendMessage(test.id, test.text)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func sendMessageHandler(w http.ResponseWriter, r *http.Request) {
	requestData, _ := io.ReadAll(r.Body)
	jsonData := &tgbot.SendMessage{}
	err := json.Unmarshal(requestData, jsonData)

	switch {
	case err != nil:
		w.WriteHeader(http.StatusBadRequest)
	case jsonData.ID < 0:
		w.WriteHeader(http.StatusBadRequest)
	case jsonData.Text == returnIncorectJSON:
		w.WriteHeader(http.StatusOK)
		err := w.Header().Write(bytes.NewBuffer([]byte("Специально отправляю не верные данные")))

		if err != nil {
			log.Println(fmt.Errorf("При записе в тело ответа, произошла ошибка: %w", err))
		}
	default:
		w.WriteHeader(http.StatusOK)

		serverAnswer := &tgbot.DefaultServerAnswer{
			Ok: true,
		}

		err = json.NewEncoder(w).Encode(serverAnswer)

		if err != nil {
			log.Println(fmt.Errorf("При записе в тело ответа, произошла ошибка: %w", err))
		}
	}
}

func UpdatesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet || r.Method == http.MethodPost {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func MessageHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Microsecond) // создание искусственной задержки

	methodGet := r.Method == http.MethodGet
	methodPost := r.Method == http.MethodPost

	if methodGet || methodPost {
		if r.Header.Get(content) != jsonType {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func getUpdatesHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)

	response := tgbot.GetUpdateAnswer{
		DefaultServerAnswer: tgbot.DefaultServerAnswer{
			Ok: true,
		},
		Updates: []tgbot.Update{
			{UpdateID: 777,
				Msg: tgbot.Message{
					From: tgbot.User{ID: 55},
					Text: "hello"},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(response)

	if err != nil {
		log.Println(fmt.Errorf("ошибка при записи в тело ответа: %w", err))
	}
}
