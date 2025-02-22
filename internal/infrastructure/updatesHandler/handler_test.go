package updatesHandler_test

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"linkTraccer/internal/application/botService/mocks"
	"linkTraccer/internal/domain/updatesServer"
	"linkTraccer/internal/infrastructure/updatesHandler"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	updatesPath = "/updates"
)

var testJson = &updatesServer.ApiErrResponse{
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
		{
			codeHTTP: 200,
			data:     testJson,
		},
		{
			codeHTTP: 400,
			data:     nil,
		},
	}

	for _, test := range tests {
		response := httptest.NewRecorder()

		if test.data != nil {
			updatesHandler.WriteInResponse(response, test.codeHTTP, test.data) // исправить этот тест
		} else {
			updatesHandler.WriteInResponse(response, test.codeHTTP, nil)
		}

		assert.Equal(t, test.codeHTTP, response.Code)
	}
}

func TestUpdateServer_HandleLinkUpdates(t *testing.T) {
	tgClient := mocks.NewTgClient(t)

	tgClient.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

	updateHandler := updatesHandler.New(tgClient)
	client := http.Client{Timeout: time.Second * 10}
	mux := http.NewServeMux()

	mux.HandleFunc(updatesPath, updateHandler.HandleLinkUpdates)

	testServer := httptest.NewServer(mux)

	defer testServer.Close()

	byteJSON, _ := json.Marshal(validJson)

	type testCase struct {
		name    string
		method  string
		body    *bytes.Buffer
		data    []byte
		correct bool
	}

	tests := []testCase{
		{
			name:    "Тест на недопустимый метод сервера",
			method:  http.MethodGet,
			body:    nil,
			data:    []byte("Hello word"),
			correct: false,
		},
		{
			name:    "Тест на недопустимый метод сервера",
			method:  http.MethodPut,
			body:    bytes.NewBuffer(byteJSON),
			data:    byteJSON,
			correct: false,
		},
		{
			name:    "Корректный тест",
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
			req, err = http.NewRequest(test.method, testServer.URL+updatesPath, nil)
		} else {
			req, err = http.NewRequest(test.method, testServer.URL+updatesPath, test.body)
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
