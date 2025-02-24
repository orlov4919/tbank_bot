package botClient_test

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"linkTraccer/internal/domain/dto"
	"linkTraccer/internal/infrastructure/botClient"
	"linkTraccer/internal/infrastructure/botClient/mocks"
	"net/http"
	"testing"
)

var testErr = errors.New("Ошибка при выполнении запроса")

var linkUpdate = &dto.LinkUpdate{
	ID:          1,
	URL:         "stackoverflow.com",
	Description: "Тема с вопросом обновилась",
	TgChatIds:   []int{2, 8, 12, 15},
}

func TestBotClient_SendLinkUpdates(t *testing.T) {

	type TestCase struct {
		name    string
		host    string
		client  botClient.HTTPClient
		update  *dto.LinkUpdate
		correct bool
	}

	goodClient := mocks.NewHTTPClient(t)
	badClient := mocks.NewHTTPClient(t)
	errClient := mocks.NewHTTPClient(t)

	goodClient.On("Do", mock.Anything).Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(nil)}, nil)
	badClient.On("Do", mock.Anything).Return(&http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(nil)}, nil)
	errClient.On("Do", mock.Anything).Return(&http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(nil)}, testErr)

	tests := []TestCase{
		{
			name:    "Отправляем новое обновление",
			host:    "localhost:8080",
			client:  goodClient,
			update:  linkUpdate,
			correct: true,
		},
		{
			name:    "Указали неверный хост, ожидаем ошибку",
			host:    "localhost:9090",
			client:  badClient,
			update:  linkUpdate,
			correct: false,
		},
		{
			name:    "Указан верный хост, но запрос закончился с ошибкой (например по таймауту)",
			host:    "localhost:8080",
			client:  errClient,
			update:  linkUpdate,
			correct: false,
		},
	}

	for _, test := range tests {
		client := botClient.New(test.host, test.client)
		err := client.SendLinkUpdates(test.update)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}
