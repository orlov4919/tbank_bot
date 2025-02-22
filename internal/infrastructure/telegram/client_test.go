package telegram_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"linkTraccer/internal/domain/tgbot"
	"linkTraccer/internal/infrastructure/telegram"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	content            = "content-type"
	jsonType           = "application/json"
	updatesPath        = "/getUpdates"
	messagePath        = "/sendMessage"
	testBotToken       = "12345"
	randomStr          = "sdfklsjc"
	returnIncorectJSON = "верни неправильный json"
)

func TestRequestToBotAPI(t *testing.T) {

	type TestCase struct {
		name    string
		method  string
		path    string
		correct bool
		client  *http.Client
		data    io.Reader
	}

	tests := []TestCase{
		{
			name:    "Верный запрос на получение обновлений от сервера",
			method:  http.MethodGet,
			path:    updatesPath,
			client:  http.DefaultClient,
			data:    nil,
			correct: true,
		},
		{
			name:    "Не правильно указан путь для получения обновлений",
			method:  http.MethodGet,
			path:    "/Updates",
			client:  http.DefaultClient,
			data:    nil,
			correct: false,
		},

		{
			name:    "Не правильно указан метод для получения обновлений",
			method:  "//",
			path:    updatesPath,
			client:  http.DefaultClient,
			data:    nil,
			correct: false,
		},

		{
			method:  "Posts",
			path:    updatesPath,
			client:  http.DefaultClient,
			data:    nil,
			correct: false,
		},

		{
			name:    "Верный запрос для отправления сообщений",
			method:  http.MethodPost,
			path:    messagePath,
			client:  http.DefaultClient,
			data:    &bytes.Buffer{},
			correct: true,
		},

		{
			name:    "Тест на таймаут при запросе",
			method:  http.MethodPost,
			path:    messagePath,
			client:  &http.Client{Timeout: time.Nanosecond},
			data:    &bytes.Buffer{},
			correct: false,
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/getUpdates", UpdatesHandler)
	mux.HandleFunc("/sendMessage", MessageHandler)

	server := httptest.NewServer(mux)

	defer server.Close()

	for _, test := range tests {
		parsedURL, _ := url.Parse(server.URL + test.path)

		_, err := telegram.RequestToAPI(test.client, parsedURL, test.method, test.data)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestTgClient_HandleUsersUpdates(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/bot"+testBotToken+"/getUpdates", getUpdatesHandler)

	server := httptest.NewServer(mux)

	defer server.Close()

	serverURL, _ := url.Parse(server.URL)
	serverHost := serverURL.Host

	type TestCase struct {
		name    string
		client  *telegram.TgClient
		offset  int
		limit   int
		correct bool
	}

	testCases := []TestCase{
		{
			name:    "Передаем при создании клиента, верный хост сервера",
			client:  telegram.NewClient(serverHost),
			offset:  5,
			limit:   100,
			correct: true,
		},
		{
			name:    "Передаем при создании клиента, неверный хост сервера",
			client:  telegram.NewClient(serverHost + randomStr),
			offset:  5,
			limit:   100,
			correct: false,
		},
		{
			name:    "Передаем отрицательный лимит при запросе",
			client:  telegram.NewClient(serverHost + randomStr),
			offset:  5,
			limit:   -2,
			correct: false,
		},
	}

	for _, test := range testCases {

		test.client.ChangeSchemeHTTP()
		test.client.SetBotToken(testBotToken)

		_, err := test.client.HandleUsersUpdates(test.offset, test.limit)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestTgClient_SendMessage(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	defer server.Close()

	mux.HandleFunc("/bot"+testBotToken+"/sendMessage", sendMessageHandler)

	serverURL, _ := url.Parse(server.URL)
	serverHost := serverURL.Host

	type TestCase struct {
		name    string
		id      int
		text    string
		bot     *telegram.TgClient
		correct bool
	}

	testCases := []TestCase{

		{
			name:    "Корректная отправка сообщения",
			id:      15,
			text:    "Привет мир",
			bot:     telegram.NewClient(serverHost),
			correct: true,
		},
		{
			name:    "Передаем отрицателный id",
			id:      -225,
			text:    "Привет мир",
			bot:     telegram.NewClient(serverHost),
			correct: false,
		},
		{
			name:    "С помощью текста запроса, просим сервер что бы вернулся не тот json, который ожидается",
			id:      228,
			text:    returnIncorectJSON,
			bot:     telegram.NewClient(serverHost),
			correct: false,
		},
	}

	for _, test := range testCases {

		test.bot.SetBotToken(testBotToken)
		test.bot.ChangeSchemeHTTP()

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
	jsonData := &telegram.SendMessage{}
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

		serverAnswer := &telegram.DefaultServerAnswer{
			Ok: true,
		}

		err = json.NewEncoder(w).Encode(serverAnswer)

		if err != nil {
			log.Println(fmt.Errorf("При записе в тело ответа, произошла ошибка: %w", err))
		}
	}
}

func getUpdatesHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)

	response := telegram.GetUpdateAnswer{
		DefaultServerAnswer: telegram.DefaultServerAnswer{
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

func UpdatesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet || r.Method == http.MethodPost {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func MessageHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Microsecond)

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
