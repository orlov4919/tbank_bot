package stackoverflow_test

import (
	"bytes"
	"errors"
	"io"
	"linkTraccer/internal/domain/scrapper"
	"linkTraccer/internal/infrastructure/siteclients/mocks"
	"linkTraccer/internal/infrastructure/siteclients/stackoverflow"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	host      = "api.stackexchange.com"
	wrongJSON = "Hello World"
	goodJSON  = `{
					"items" : [{"title" : "hello"}]
                }`
)

var (
	errNet = errors.New("ошибка при выполнении запроса")
)

func TestStackClient_StaticLinkCheck(t *testing.T) {
	type TestCase struct {
		name   string
		link   string
		result bool
	}

	mockedHTTPClient := mocks.NewHTTPClient(t)
	client := stackoverflow.NewClient(host, mockedHTTPClient, stackoverflow.HTMLStrCleaner(200))

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
			link:   "https://stackoverflow.com/questions/76814302/embedded-linux-flash-storage-security/15",
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
	clientWithErr := mocks.NewHTTPClient(t)

	clientToGoodLinks.On("Do", mock.Anything).Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(nil)}, nil)
	clientToWrongLinks.On("Do", mock.Anything).Return(&http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(nil)}, nil)
	clientWithErr.On("Do", mock.Anything).Return(&http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(nil)}, errNet)

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
		{
			name:   "Произошла сетевая ошибка",
			link:   "https://stackoverflow.com/questions/76814302",
			client: clientWithErr,
			result: false,
		},
		{
			name:   "Ссылку невозможно распарсить",
			link:   "https://stackoverflow.com/questions\n/76814302/",
			client: clientToWrongLinks,
			result: false,
		},
	}

	for _, test := range tests {
		client := stackoverflow.NewClient(host, test.client, stackoverflow.HTMLStrCleaner(200))
		actualRes := client.CanTrack(test.link)

		assert.Equal(t, test.result, actualRes)
	}
}

func TestStackClient_NewUpdate(t *testing.T) {
	type TestCase struct {
		name        string
		req         *http.Request
		client      stackoverflow.HTTPClient
		update      any
		expectedRes any
		correct     bool
	}

	clientWithOkStatus := mocks.NewHTTPClient(t)
	clientWithOkStatusBadJSON := mocks.NewHTTPClient(t)
	clientWithErrStatus := mocks.NewHTTPClient(t)
	clientWithErr := mocks.NewHTTPClient(t)

	clientWithOkStatus.On("Do", mock.Anything).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBuffer([]byte(goodJSON)))}, nil)
	clientWithOkStatusBadJSON.On("Do", mock.Anything).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBuffer([]byte(wrongJSON)))}, nil)
	clientWithErrStatus.On("Do", mock.Anything).
		Return(&http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(nil)}, nil)
	clientWithErr.On("Do", mock.Anything).
		Return(&http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(nil)}, errNet)

	tests := []TestCase{
		{
			name: "Произошла сетевая ошибка при получении обновлений",
			req: &http.Request{
				URL: &url.URL{Path: "questions"},
			},
			client:      clientWithErr,
			update:      &scrapper.StackAnswers{},
			correct:     false,
			expectedRes: &scrapper.StackAnswers{},
		},
		{
			name: "Пришел код ответа отличный от 200",
			req: &http.Request{
				URL: &url.URL{Path: "questions"},
			},
			client:      clientWithErrStatus,
			update:      &scrapper.StackAnswers{},
			correct:     false,
			expectedRes: &scrapper.StackAnswers{},
		},
		{
			name: "Пришел невалидный JSON",
			req: &http.Request{
				URL: &url.URL{Path: "questions"},
			},
			client:      clientWithOkStatusBadJSON,
			update:      &scrapper.StackAnswers{},
			correct:     false,
			expectedRes: &scrapper.StackAnswers{},
		},
		{
			name: "Пришел валидный JSON",
			req: &http.Request{
				URL: &url.URL{Path: "questions"},
			},
			client:  clientWithOkStatus,
			update:  &scrapper.StackAnswers{},
			correct: true,
			expectedRes: &scrapper.StackAnswers{
				Items: []scrapper.StackAnswer{{Title: "hello"}},
			},
		},
	}

	for _, test := range tests {
		client := stackoverflow.NewClient(host, test.client, stackoverflow.HTMLStrCleaner(200))
		err := client.NewUpdate(test.req, test.update)

		if test.correct {
			assert.Equal(t, test.expectedRes, test.update)
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestStackClient_NewAnswers(t *testing.T) {
	clientWithOkStatus := mocks.NewHTTPClient(t)
	clientWithErrStatus := mocks.NewHTTPClient(t)

	clientWithOkStatus.On("Do", mock.Anything).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBuffer([]byte(goodJSON)))}, nil)
	clientWithErrStatus.On("Do", mock.Anything).
		Return(&http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(nil)}, nil)

	type TestCase struct {
		name        string
		questionID  string
		client      stackoverflow.HTTPClient
		expectedRes any
		correct     bool
	}

	tests := []TestCase{
		{
			name:        "Ошибка при получении обновлений",
			questionID:  "1",
			client:      clientWithErrStatus,
			expectedRes: &scrapper.StackAnswers{},
			correct:     false,
		},
		{
			name:       "Получение новых ответов, без ошибок",
			questionID: "1",
			client:     clientWithOkStatus,
			expectedRes: &scrapper.StackAnswers{
				Items: []scrapper.StackAnswer{{Title: "hello"}}},
			correct: true,
		},
	}

	for _, test := range tests {
		client := stackoverflow.NewClient(host, test.client, stackoverflow.HTMLStrCleaner(200))
		updates, err := client.NewAnswers(test.questionID, time.Now())

		if test.correct {
			assert.Equal(t, test.expectedRes, updates)
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestStackClient_NewComments(t *testing.T) {
	clientWithOkStatus := mocks.NewHTTPClient(t)
	clientWithErrStatus := mocks.NewHTTPClient(t)

	clientWithOkStatus.On("Do", mock.Anything).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBuffer([]byte(goodJSON)))}, nil)
	clientWithErrStatus.On("Do", mock.Anything).
		Return(&http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(nil)}, nil)

	type TestCase struct {
		name        string
		questionID  string
		client      stackoverflow.HTTPClient
		expectedRes any
		correct     bool
	}

	tests := []TestCase{
		{
			name:        "Ошибка при получении обновлений",
			questionID:  "1",
			client:      clientWithErrStatus,
			expectedRes: &scrapper.StackComments{},
			correct:     false,
		},
		{
			name:       "Получение новых комментариев, без ошибок",
			questionID: "1",
			client:     clientWithOkStatus,
			expectedRes: &scrapper.StackComments{
				Items: []scrapper.StackComment{{Title: "hello"}}},
			correct: true,
		},
	}

	for _, test := range tests {
		client := stackoverflow.NewClient(host, test.client, stackoverflow.HTMLStrCleaner(200))
		updates, err := client.NewComments(test.questionID, time.Now())

		if test.correct {
			assert.Equal(t, test.expectedRes, updates)
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

//
// func TestStackClient_LinkState(t *testing.T) {
//	type TestCase struct {
//		name    string
//		link    string
//		client  stackoverflow.HTTPClient
//		correct bool
//	}
//
//	clientToGoodLinks := mocks.NewHTTPClient(t)
//	clientToWrongLinks := mocks.NewHTTPClient(t)
//	clientWithWrongBody := mocks.NewHTTPClient(t)
//
//	clientToGoodLinks.On("Do", mock.Anything).
//		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBuffer([]byte(jsonAnswer)))}, nil)
//	clientWithWrongBody.On("Do", mock.Anything).
//		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBuffer([]byte(wrongJSONAnswer)))}, nil)
//	clientToWrongLinks.On("Do", mock.Anything).
//		Return(&http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(nil)}, nil)
//
//	tests := []TestCase{
//		{
//			name:    "Правильная ссылка на вопрос",
//			link:    "https://stackoverflow.com/questions/76814302/embedded-linux-flash-storage-security",
//			client:  clientToGoodLinks,
//			correct: true,
//		},
//		{
//			name:    "Сcылка на вопрос с огромным Id",
//			link:    "https://stackoverflow.com/questions/9999999999/embedded-linux-flash-storage-security",
//			client:  clientToWrongLinks,
//			correct: false,
//		},
//		{
//			name:    "Ссылка на вопрос верная, но пришел json не того формата",
//			link:    "https://stackoverflow.com/questions/76814302/embedded-linux-flash-storage-security",
//			client:  clientWithWrongBody,
//			correct: false,
//		},
//		{
//			name:    "Ссылка на вопрос верная, но пришел json не того формата",
//			link:    "https://stackoverflow.com/questions/76814302/embedded-linux-flash-storage-security",
//			client:  clientWithWrongBody,
//			correct: false,
//		},
//		{
//			name:    "Ссылка на вопрос не верная",
//			link:    "https://tbank.ru/questions/76814302/embedded-linux-flash-storage-security",
//			client:  clientWithWrongBody,
//			correct: false,
//		},
//	}
//
//	for _, test := range tests {
//		client := stackoverflow.NewClient(host, test.client)
//
//		_, err := client.LinkState(test.link)
//
//		if test.correct {
//			assert.NoError(t, err)
//		} else {
//			assert.Error(t, err)
//		}
//	}
// }
