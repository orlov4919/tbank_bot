package stackoverflow_test

import (
	"bytes"
	"io"
	"linkTraccer/internal/infrastructure/siteclients/mocks"
	"linkTraccer/internal/infrastructure/siteclients/stackoverflow"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	host       = "api.stackexchange.com"
	jsonAnswer = `{
					  "items": [
						{
						  "last_activity_date": 1240232332
						}
					  ]
					}`

	wrongJSONAnswer = `{
					  "items": []
					}`
)

func TestStackClient_StaticLinkCheck(t *testing.T) {
	type TestCase struct {
		name   string
		link   string
		result bool
	}

	mockedHTTPClient := mocks.NewHTTPClient(t)
	client := stackoverflow.NewClient(host, mockedHTTPClient)

	tests := []TestCase{
		{
			name:   "Правильная ссылка на вопрос",
			link:   "https://stackoverflow.com/questions/76814302/embedded-linux-flash-storage-security",
			result: true,
		},
		{
			name:   "Ссылка с отрицательным id вопроса",
			link:   "https://stackoverflow.com/questions/-10/embedded-linux-flash-storage-security",
			result: false,
		},
		{
			name:   "Ссылка с неправильным хостом",
			link:   "https://tbank.com/questions/5000000000/embedded-linux-flash-storage-security",
			result: false,
		},
		{
			name:   "Ссылка с неправильной схемой",
			link:   "http://stackoverflow.com/questions/5000000000/embedded-linux-flash-storage-security",
			result: false,
		},
		{
			name:   "Ссылка с неправильным путем",
			link:   "http://stackoverflow.com/question/1234567/embedded-linux-flash-storage-security",
			result: false,
		},
		{
			name:   "Очень длинная ссылка",
			link:   "https://stackoverflow.com/questions/76814302/embedded-linux-flash-storage-security/15/14/12",
			result: false,
		},
		{
			name:   "Сокращенная ссылка на вопрос",
			link:   "https://stackoverflow.com/questions/76814302",
			result: true,
		},
		{
			name:   "Вместо id вопроса указываем рандомную строку",
			link:   "https://stackoverflow.com/questions/abcd",
			result: false,
		},
		{
			name:   "Перед questions, добавляем несуществуюищй параметр пути",
			link:   "https://stackoverflow.com/new/questions/123456",
			result: false,
		},
		{
			name:   "Не указан id вопроса",
			link:   "https://stackoverflow.com/questions",
			result: false,
		},
	}

	for ind, test := range tests {
		parsedLink, _ := url.Parse(test.link)
		splitedLink := strings.Split(parsedLink.Path, "/")

		assert.Equal(t, test.result, client.StaticLinkCheck(parsedLink, splitedLink), ind)
	}
}

func TestStackClient_CanTrack(t *testing.T) {
	type TestCase struct {
		name   string
		link   string
		client stackoverflow.HTTPClient
		result bool
	}

	clientToGoodLinks := mocks.NewHTTPClient(t)
	clientToWrongLinks := mocks.NewHTTPClient(t)

	clientToGoodLinks.On("Do", mock.Anything).Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(nil)}, nil)
	clientToWrongLinks.On("Do", mock.Anything).Return(&http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(nil)}, nil)

	tests := []TestCase{
		{
			name:   "Правильная ссылка на вопрос",
			link:   "https://stackoverflow.com/questions/76814302/embedded-linux-flash-storage-security",
			client: clientToGoodLinks,
			result: true,
		},
		{
			name:   "Ссылка с большим id вопроса",
			link:   "https://stackoverflow.com/questions/999999999/embedded-linux-flash-storage-security",
			client: clientToWrongLinks,
			result: false,
		},
		{
			name:   "Ссылка с неправильной схемой",
			link:   "http://stackoverflow.com/questions/5000000000/embedded-linux-flash-storage-security",
			client: clientToWrongLinks,
			result: false,
		},
		{
			name:   "Сокращенная ссылка на вопрос",
			link:   "https://stackoverflow.com/questions/76814302",
			client: clientToGoodLinks,
			result: true,
		},
	}

	for _, test := range tests {
		client := stackoverflow.NewClient(host, test.client)
		actualRes := client.CanTrack(test.link)

		assert.Equal(t, test.result, actualRes)
	}
}

func TestStackClient_LinkState(t *testing.T) {
	type TestCase struct {
		name    string
		link    string
		client  stackoverflow.HTTPClient
		correct bool
	}

	clientToGoodLinks := mocks.NewHTTPClient(t)
	clientToWrongLinks := mocks.NewHTTPClient(t)
	clientWithWrongBody := mocks.NewHTTPClient(t)

	clientToGoodLinks.On("Do", mock.Anything).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBuffer([]byte(jsonAnswer)))}, nil)
	clientWithWrongBody.On("Do", mock.Anything).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBuffer([]byte(wrongJSONAnswer)))}, nil)
	clientToWrongLinks.On("Do", mock.Anything).
		Return(&http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(nil)}, nil)

	tests := []TestCase{
		{
			name:    "Правильная ссылка на вопрос",
			link:    "https://stackoverflow.com/questions/76814302/embedded-linux-flash-storage-security",
			client:  clientToGoodLinks,
			correct: true,
		},
		{
			name:    "Сcылка на вопрос с огромным Id",
			link:    "https://stackoverflow.com/questions/9999999999/embedded-linux-flash-storage-security",
			client:  clientToWrongLinks,
			correct: false,
		},
		{
			name:    "Ссылка на вопрос верная, но пришел json не того формата",
			link:    "https://stackoverflow.com/questions/76814302/embedded-linux-flash-storage-security",
			client:  clientWithWrongBody,
			correct: false,
		},
		{
			name:    "Ссылка на вопрос верная, но пришел json не того формата",
			link:    "https://stackoverflow.com/questions/76814302/embedded-linux-flash-storage-security",
			client:  clientWithWrongBody,
			correct: false,
		},
		{
			name:    "Ссылка на вопрос не верная",
			link:    "https://tbank.ru/questions/76814302/embedded-linux-flash-storage-security",
			client:  clientWithWrongBody,
			correct: false,
		},
	}

	for _, test := range tests {
		client := stackoverflow.NewClient(host, test.client)

		_, err := client.LinkState(test.link)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}
