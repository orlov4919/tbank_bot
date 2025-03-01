package scrapperhandlers_test

import (
	"errors"
	"linkTraccer/internal/infrastructure/scrapperhandlers"
	"linkTraccer/internal/infrastructure/scrapperhandlers/mocks"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	firstUserID  = 1
	secondUserID = 2
)

var errTest = errors.New("ошибка при удалении пользователя")

func TestChatHandler_HandleChatChanges(t *testing.T) {
	userRepo := mocks.NewUserRepo(t)

	userRepo.On("RegUser", mock.Anything).Return(nil)
	userRepo.On("UserExist", firstUserID).Return(true)
	userRepo.On("UserExist", secondUserID).Return(false)
	userRepo.On("DeleteUser", firstUserID).Return(nil)
	userRepo.On("DeleteUser", secondUserID).Return(errTest)

	scrapHandler := scrapperhandlers.NewChatHandler(userRepo)

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
			name: "Передаем запрос с неправильным id",
			r: &http.Request{
				Method: http.MethodPost,
				URL:    &url.URL{Path: "/tg-chat/chat"},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Передаем запрос с отрицательным id",
			r:              &http.Request{Method: http.MethodPost, URL: &url.URL{Path: "/tg-chat/-15"}},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Пытаемся добавить уже добавленного пользователя",
			r:              &http.Request{Method: http.MethodPost, URL: &url.URL{Path: "/tg-chat/1"}},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Пытаемся добавить не добавленного пользователя",
			r:              &http.Request{Method: http.MethodPost, URL: &url.URL{Path: "/tg-chat/2"}},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Пытаемся удалить добавленного пользователя",
			r:              &http.Request{Method: http.MethodDelete, URL: &url.URL{Path: "/tg-chat/1"}},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Пытаемся удалить не добавленного пользователя",
			r:              &http.Request{Method: http.MethodDelete, URL: &url.URL{Path: "/tg-chat/2"}},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()
		scrapHandler.HandleChatChanges(w, test.r)

		assert.Equal(t, test.expectedStatus, w.Code)
	}
}
