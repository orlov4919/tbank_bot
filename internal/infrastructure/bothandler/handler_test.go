package bothandler_test

import (
	"bytes"
	"encoding/json"
	"io"
	"linkTraccer/internal/application/botService/mocks"
	"linkTraccer/internal/domain/dto"
	"linkTraccer/internal/infrastructure/bothandler"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var randomStrBytes = []byte("hello word")

var testJSON = &dto.ApiErrResponse{
	Description: "ошибка для тестирования",
	Code:        "400",
}

var linkUpdate = &dto.LinkUpdate{
	ID:          1,
	URL:         "google.com",
	Description: "Новое уведомление",
	TgChatIds:   []int{1, 2},
}

func TestUpdateServer_WriteResponseData(t *testing.T) {
	type TestCase struct {
		codeHTTP int
		data     any
	}

	tests := []TestCase{
		{
			codeHTTP: 200,
			data:     nil,
		},
		{
			codeHTTP: 400,
			data:     testJSON,
		},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()

		bothandler.WriteInResponse(w, test.codeHTTP, test.data)
		assert.Equal(t, test.codeHTTP, w.Code)

		if test.data != nil {
			respJSON := &dto.ApiErrResponse{}
			respData, _ := io.ReadAll(w.Body)

			_ = json.Unmarshal(respData, respJSON)
			assert.Equal(t, test.data, respJSON)
		}
	}
}

func TestUpdateServer_HandleLinkUpdates(t *testing.T) {
	tgClient := mocks.NewTgClient(t)

	tgClient.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

	botHandler := bothandler.New(tgClient)
	linkUpdateJSON, _ := json.Marshal(linkUpdate)

	type testCase struct {
		name       string
		r          *http.Request
		httpStatus int
	}

	tests := []testCase{
		{
			name:       "Тест на недопустимый метод сервера",
			r:          &http.Request{Method: http.MethodGet},
			httpStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "Тест на передачу не валидных данных",
			r:          &http.Request{Method: http.MethodPost, Body: io.NopCloser(bytes.NewBuffer(randomStrBytes))},
			httpStatus: http.StatusBadRequest,
		},
		{
			name:       "Тест на передачу валидных данных",
			r:          &http.Request{Method: http.MethodPost, Body: io.NopCloser(bytes.NewBuffer(linkUpdateJSON))},
			httpStatus: http.StatusOK,
		},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()

		botHandler.HandleLinkUpdates(w, test.r)

		assert.Equal(t, test.httpStatus, w.Code)
	}
}
