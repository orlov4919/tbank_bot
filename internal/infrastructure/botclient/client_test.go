package botclient_test

import (
	"errors"
	"io"
	"linkTraccer/internal/domain/dto"
	"linkTraccer/internal/infrastructure/botclient"
	"linkTraccer/internal/infrastructure/botclient/mocks"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var errTest = errors.New("ошибка при выполнении запроса")

var linkUpdate = &dto.LinkUpdate{
	ID:          1,
	URL:         "stackoverflow.com",
	Description: "Тема с вопросом обновилась",
	TgChatIDs:   []int{2, 8, 12, 15},
}

func TestBotClient_SendLinkUpdates(t *testing.T) {
	type TestCase struct {
		name    string
		host    string
		client  botclient.HTTPClient
		update  *dto.LinkUpdate
		correct bool
	}

	goodClient := mocks.NewHTTPClient(t)
	badClient := mocks.NewHTTPClient(t)
	errClient := mocks.NewHTTPClient(t)

	goodClient.On("Do", mock.Anything).Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(nil)}, nil)
	badClient.On("Do", mock.Anything).Return(&http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(nil)}, nil)
	errClient.On("Do", mock.Anything).Return(&http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(nil)}, errTest)

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
		client := botclient.New(test.host, test.client)
		err := client.SendLinkUpdates(test.update)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}
