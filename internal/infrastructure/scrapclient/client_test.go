package scrapclient_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"linkTraccer/internal/domain/scrapper"
	"linkTraccer/internal/domain/tgbot"
	"linkTraccer/internal/infrastructure/scrapclient"
	"linkTraccer/internal/infrastructure/scrapclient/mocks"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	host                 = "localhost"
	port                 = ":8080"
	randomStr            = "Hello word"
	savedLink tgbot.Link = "tbank.ru"
)

var links = &scrapper.ListLinksResponse{
	Size: 1,
	Links: []scrapper.LinkResponse{
		{
			URL:     savedLink,
			Tags:    []string{},
			Filters: []string{},
		},
	},
}

var linksJSON, _ = json.Marshal(links)

var errTest = errors.New("ошибка для теста")

var badResponse = &http.Response{StatusCode: http.StatusBadRequest,
	Body: io.NopCloser(bytes.NewBuffer([]byte{}))}

var goodResponse = &http.Response{StatusCode: http.StatusOK,
	Body: io.NopCloser(bytes.NewBuffer(linksJSON))}

var badBodyResponse = &http.Response{StatusCode: http.StatusOK,
	Body: io.NopCloser(bytes.NewBuffer([]byte(randomStr)))}

func TestScrapperClient_RegUser(t *testing.T) {
	badClient := mocks.NewHTTPClient(t)
	badRequestClient := mocks.NewHTTPClient(t)
	goodClient := mocks.NewHTTPClient(t)

	badClient.On("Do", mock.Anything).Return(nil, errTest)
	badRequestClient.On("Do", mock.Anything).Return(badResponse, nil)
	goodClient.On("Do", mock.Anything).Return(goodResponse, nil)

	type testCase struct {
		name    string
		client  scrapclient.HTTPClient
		id      tgbot.ID
		correct bool
	}

	tests := []testCase{
		{
			name:    "ошибка во время выполнения запроса",
			client:  badClient,
			id:      10,
			correct: false,
		},
		{
			name:    "ошибка неправильного запроса",
			client:  badRequestClient,
			id:      -10,
			correct: false,
		},
		{
			name:    "правильный запрос",
			client:  goodClient,
			id:      1,
			correct: true,
		},
	}

	for _, test := range tests {
		client := scrapclient.New(test.client, host, port)
		err := client.RegUser(test.id)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestScrapperClient_UserLinks(t *testing.T) {
	badClient := mocks.NewHTTPClient(t)
	badRequestClient := mocks.NewHTTPClient(t)
	badBodyClient := mocks.NewHTTPClient(t)
	goodClient := mocks.NewHTTPClient(t)

	badClient.On("Do", mock.Anything).Return(nil, errTest)
	badRequestClient.On("Do", mock.Anything).Return(badResponse, nil)
	badBodyClient.On("Do", mock.Anything).Return(badBodyResponse, nil)
	goodClient.On("Do", mock.Anything).Return(goodResponse, nil)

	type testCase struct {
		name    string
		client  scrapclient.HTTPClient
		id      tgbot.ID
		links   []tgbot.Link
		correct bool
	}

	tests := []testCase{
		{
			name:    "ошибка во время выполнения запроса",
			client:  badClient,
			id:      10,
			links:   nil,
			correct: false,
		},
		{
			name:    "пришла ошибка от сервера",
			client:  badRequestClient,
			id:      10,
			links:   nil,
			correct: false,
		},
		{
			name:    "в качестве ответа в теле пришел не json",
			client:  badBodyClient,
			id:      10,
			links:   nil,
			correct: false,
		},
		{
			name:    "тест без ошибок",
			client:  goodClient,
			id:      10,
			links:   []tgbot.Link{savedLink},
			correct: true,
		},
	}

	for _, test := range tests {
		client := scrapclient.New(test.client, host, port)
		links, err := client.UserLinks(test.id)

		if test.correct {
			assert.ElementsMatch(t, test.links, links)
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestScrapperClient_RemoveLink(t *testing.T) {
	badClient := mocks.NewHTTPClient(t)
	badRequestClient := mocks.NewHTTPClient(t)
	goodClient := mocks.NewHTTPClient(t)

	badClient.On("Do", mock.Anything).Return(nil, errTest)
	badRequestClient.On("Do", mock.Anything).Return(badResponse, nil)
	goodClient.On("Do", mock.Anything).Return(goodResponse, nil)

	type testCase struct {
		name    string
		client  scrapclient.HTTPClient
		link    tgbot.Link
		id      tgbot.ID
		correct bool
	}

	tests := []testCase{
		{
			name:    "ошибка во время выполнения запроса",
			client:  badClient,
			id:      10,
			link:    savedLink,
			correct: false,
		},
		{
			name:    "ошибка от сервера",
			client:  badRequestClient,
			id:      10,
			link:    savedLink,
			correct: false,
		},
		{
			name:    "правильный запрос",
			client:  goodClient,
			id:      1,
			correct: true,
		},
	}

	for _, test := range tests {
		client := scrapclient.New(test.client, host, port)
		err := client.RemoveLink(test.id, test.link)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestScrapperClient_AddLink(t *testing.T) {
	badClient := mocks.NewHTTPClient(t)
	badRequestClient := mocks.NewHTTPClient(t)
	goodClient := mocks.NewHTTPClient(t)

	badClient.On("Do", mock.Anything).Return(nil, errTest)
	badRequestClient.On("Do", mock.Anything).Return(badResponse, nil)
	goodClient.On("Do", mock.Anything).Return(goodResponse, nil)

	type testCase struct {
		name    string
		client  scrapclient.HTTPClient
		data    *tgbot.ContextData
		id      tgbot.ID
		correct bool
	}

	tests := []testCase{
		{
			name:    "ошибка во время выполнения запроса",
			client:  badClient,
			id:      10,
			data:    &tgbot.ContextData{},
			correct: false,
		},
		{
			name:    "ошибка от сервера",
			client:  badRequestClient,
			id:      10,
			data:    &tgbot.ContextData{},
			correct: false,
		},
		{
			name:    "ошибка от сервера",
			client:  goodClient,
			id:      10,
			data:    &tgbot.ContextData{},
			correct: true,
		},
	}

	for _, test := range tests {
		client := scrapclient.New(test.client, host, port)
		err := client.AddLink(test.id, test.data)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}
