package scrapperHandlers_test

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/assert"
	"io"
	"linkTraccer/internal/infrastructure/scrapperHandlers"
	"linkTraccer/internal/infrastructure/scrapperHandlers/mocks"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	firstUserId  = 1
	secondUserId = 2
)

var testErr = errors.New("Ошибка при удалении пользователя")
var wrongData = []byte("abc")
var negativeID = []byte("-15")
var firstId = []byte("1")
var secondID = []byte("2")

func TestChatHandler_HandleChatChanges(t *testing.T) {
	userRepo := mocks.NewUserRepo(t)

	userRepo.On("UserExist", firstUserId).Return(true)
	userRepo.On("UserExist", secondUserId).Return(false)
	userRepo.On("DeleteUser", firstUserId).Return(nil)
	userRepo.On("DeleteUser", secondUserId).Return(testErr)

	scrapHandler := scrapperHandlers.NewChatHandler(userRepo)

	type testCase struct {
		name           string
		r              *http.Request
		expectedStatus int
	}

	tests := []testCase{
		{
			name:           "Передаем запрос с не поддерживаемым методом",
			r:              &http.Request{Method: http.MethodPut},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Передаем запрос с неправильным body",
			r:              &http.Request{Method: http.MethodPost, Body: io.NopCloser(bytes.NewBuffer(wrongData))},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Передаем запрос с отрицательным id",
			r:              &http.Request{Method: http.MethodPost, Body: io.NopCloser(bytes.NewBuffer(negativeID))},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Пытаемся добавить уже добавленного пользователя",
			r:              &http.Request{Method: http.MethodPost, Body: io.NopCloser(bytes.NewBuffer(firstId))},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Пытаемся добавить не добавленного пользователя",
			r:              &http.Request{Method: http.MethodPost, Body: io.NopCloser(bytes.NewBuffer(secondID))},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Пытаемся удалить добавленного пользователя",
			r:              &http.Request{Method: http.MethodDelete, Body: io.NopCloser(bytes.NewBuffer(firstId))},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Пытаемся удалить не добавленного пользователя",
			r:              &http.Request{Method: http.MethodDelete, Body: io.NopCloser(bytes.NewBuffer(secondID))},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()
		scrapHandler.HandleChatChanges(w, test.r)

		assert.Equal(t, test.expectedStatus, w.Code)
	}
}
