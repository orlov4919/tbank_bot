package updatesServer_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"linkTraccer/internal/domain/tgbot/mocks"
	"linkTraccer/internal/domain/updatesServer"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	updatesPath = "/updates"
	randomData  = "abcdsfjklj"
)

const(
	serverUrl := http://localhost:8080/page
)

var testJson = &updatesServer.ApiErrorResponse{
	Description: "ошибка для тестирования",
	Code:        "400",
}

var validJson = &updatesServer.LinkUpdate{
	ID:          15,
	URL:         "google.com",
	Description: "Новое уведомление",
	TgChatIds:   []int{1, 2},
}

func TestUpdateServer_WriteResponseData(t *testing.T) {
	type TestCase struct {
		codeHTTP int
		data     any
		correct  bool
	}

	tests := []TestCase{
		{codeHTTP: 200,
			data:    testJson,
			correct: true,
		},
		{
			codeHTTP: 400,
			data:     nil,
			correct:  false,
		},
		//{
		//	codeHTTP: 200,
		//	data:     "abcd",
		//	correct:  true,
		//},
	}

	for _, test := range tests {
		response := httptest.NewRecorder()

		if test.data != nil {
			updatesServer.WriteInResponse(response, test.codeHTTP, test.data) // исправить этот тест
		} else {
			updatesServer.WriteInResponse(response, test.codeHTTP, nil)
		}

		assert.Equal(t, test.codeHTTP, response.Code)
	}
}

func TestUpdateServer_HandleLinkUpdates(t *testing.T) {
	client := http.Client{Timeout: time.Second * 10}
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	tgBot := mocks.NewBotService(t)

	tgBot.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

	defer server.Close()

	updateServer := updatesServer.New(server.URL, tgBot)

	mux.HandleFunc(updatesPath, updateServer.HandleLinkUpdates)

	byteJSON, _ := json.Marshal(validJson)

	type testCase struct {
		method  string
		body    *bytes.Buffer
		data    []byte
		correct bool
	}

	tests := []testCase{{
		method:  http.MethodGet,
		body:    nil,
		data:    []byte("Hello word"),
		correct: false,
	}, {
		method:  http.MethodGet,
		body:    bytes.NewBuffer(byteJSON),
		data:    byteJSON,
		correct: false,
	},
		{
			method:  http.MethodPost,
			body:    bytes.NewBuffer(byteJSON),
			data:    byteJSON,
			correct: true,
		},
	}

	var err error
	var req *http.Request
	var resp *http.Response

	for _, test := range tests {

		if test.body == nil {
			req, err = http.NewRequest(test.method, server.URL+updatesPath, nil)
		} else {
			req, err = http.NewRequest(test.method, server.URL+updatesPath, test.body)
		}

		if err != nil {
			log.Println("при создании запроса, возникла ошибка")
			continue
		}

		resp, err = client.Do(req)

		if err != nil {
			log.Println("ошибка при выполнении запроса клиентом")
		}

		if test.correct {
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		} else {
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		}
	}
}

//func TestUpdateServer_StartUpdatesService(t *testing.T) {
//	//client := http.Client{Timeout: time.Second * 10}
//
//	tgBot := mocks.NewBotService(t)
//
//
//	//tgBot.On("SendMessage", mock.Anything, mock.Anything).Return(nil)
//
//	updateServer := updatesServer.New(server.URL, tgBot)
//
//
//}
