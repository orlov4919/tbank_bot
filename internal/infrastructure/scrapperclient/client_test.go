package scrapperclient_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"linkTraccer/internal/domain/scrapper"
	"linkTraccer/internal/domain/tgbot"
	"linkTraccer/internal/infrastructure/scrapperclient"
	"linkTraccer/internal/infrastructure/scrapperclient/mocks"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type HTTPClient = scrapperclient.HTTPClient
type ID = tgbot.ID
type Link = tgbot.Link
type ListLinksResponse = scrapper.ListLinksResponse
type LinkResponse = scrapper.LinkResponse

const localhost = "localhost:8080"
const randomStr = "Hello word"
const savedLink Link = "tbank.ru"

var links = &ListLinksResponse{
	Size: 1,
	Links: []LinkResponse{
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
		client  HTTPClient
		id      ID
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
		client := scrapperclient.New(test.client, localhost)
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
		client  HTTPClient
		id      ID
		links   []Link
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
			links:   []Link{savedLink},
			correct: true,
		},
	}

	for _, test := range tests {
		client := scrapperclient.New(test.client, localhost)
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
		client  HTTPClient
		link    Link
		id      ID
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
		client := scrapperclient.New(test.client, localhost)
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
		client  HTTPClient
		data    *tgbot.ContextData
		id      ID
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
		client := scrapperclient.New(test.client, localhost)
		err := client.AddLink(test.id, test.data)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}
