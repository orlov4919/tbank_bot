package github_test

import (
	"bytes"
	"errors"
	"io"
	"linkTraccer/internal/domain/scrapper"
	"linkTraccer/internal/infrastructure/siteclients/github"
	"linkTraccer/internal/infrastructure/siteclients/mocks"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testHost  = "api.github.com"
	testToken = "123456789"
)

var (
	jsonData   = []byte(`{"items" : [{ "created_at" : "2025-02-25T11:39:14Z"}]}`)
	randomData = []byte("abcdsdfsdf")
	errTest    = errors.New("произошел таймаут")
)

type Link = scrapper.Link

func TestGitClient_StaticLinkCheck(t *testing.T) {
	mockClient := mocks.NewHTTPClient(t)
	gitClient := github.NewClient("github.com", "12345678", mockClient)

	type testCase struct {
		name    string
		link    Link
		correct bool
	}

	tests := []testCase{
		{
			name:    "Не правильно указан хост gitHub",
			link:    "https://gitehube.com/orlov4919/test",
			correct: false,
		},
		{
			name:    "Не правильно указана схема",
			link:    "http://github.com/orlov4919/test",
			correct: false,
		},
		{
			name:    "Не указан репозиторий",
			link:    "https://github.com/orlov4919/",
			correct: false,
		},
		{
			name:    "Слишком длинная ссылка",
			link:    "https://github.com/orlov4919/repoNew/123",
			correct: false,
		},
		{
			name:    "Корректная ссылка",
			link:    "https://github.com/orlov4919/test",
			correct: true,
		},
		{
			name:    "Не указан репозиторий",
			link:    "https://github.com//",
			correct: false,
		},
		{
			name:    "Слишком длинная ссылка",
			link:    "https://github.com/orlov4919/test/issues",
			correct: false,
		},
	}

	for _, test := range tests {
		parsedLink, _ := url.Parse(test.link)
		pathArgs := strings.Split(parsedLink.Path, "/")

		assert.Equal(t, test.correct, gitClient.StaticLinkCheck(parsedLink, pathArgs))
	}
}

func TestGitClient_CanTrack(t *testing.T) {
	clientWith404 := mocks.NewHTTPClient(t)
	clientWithErr := mocks.NewHTTPClient(t)
	clientWithOK := mocks.NewHTTPClient(t)

	clientWith404.On("Do", mock.Anything).Return(&http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(nil)}, nil)
	clientWithErr.On("Do", mock.Anything).Return(nil, errTest)
	clientWithOK.On("Do", mock.Anything).Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBuffer(jsonData))}, nil)

	type testCase struct {
		name    string
		link    Link
		client  github.HTTPClient
		correct bool
	}

	tests := []testCase{
		{
			name:    "Ошибка при парсинге ссылки",
			link:    "\nhttps://gitehube.com/orlov4919/test",
			client:  clientWithOK,
			correct: false,
		},
		{
			name:    "Передаем ссылку с неправильным хостом",
			link:    "https://gitehube.com/orlov4919/test",
			client:  clientWith404,
			correct: false,
		},
		{
			name:    "Передаем клиента, который таймаутит",
			link:    "https://github.com/orlov4919/test",
			client:  clientWithErr,
			correct: false,
		},
		{
			name:    "Пытаемся отследить репозиторий, которого нет",
			link:    "https://github.com/orlov4919/test1234",
			client:  clientWith404,
			correct: false,
		},
		{
			name:    "Отслеживаем существующий репозиторий",
			link:    "https://github.com/orlov4919/test",
			client:  clientWithOK,
			correct: true,
		},
	}

	for _, test := range tests {
		gitClient := github.NewClient(testHost, testToken, test.client)

		assert.Equal(t, test.correct, gitClient.CanTrack(test.link))
	}
}

func TestGitClient_LinkUpdates(t *testing.T) {
	clientWith404 := mocks.NewHTTPClient(t)
	clientWithErr := mocks.NewHTTPClient(t)
	clientWithWrongJSON := mocks.NewHTTPClient(t)
	clientWithOK := mocks.NewHTTPClient(t)

	clientWith404.On("Do", mock.Anything).
		Return(&http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(nil)}, nil)
	clientWithErr.On("Do", mock.Anything).
		Return(nil, errTest)
	clientWithWrongJSON.On("Do", mock.Anything).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBuffer(randomData))}, nil)
	clientWithOK.On("Do", mock.Anything).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBuffer(jsonData))}, nil)

	type testCase struct {
		name    string
		link    Link
		client  github.HTTPClient
		correct bool
		updates scrapper.LinkUpdates
	}

	tests := []testCase{
		{
			name:    "Произошла ошибка при парсинге ссылки",
			link:    "\nhttps://gitehube.com/orlov4919/test",
			client:  clientWith404,
			correct: false,
		},
		{
			name:    "Передаем ссылку с неправильным хостом",
			link:    "https://gitehube.com/orlov4919/test",
			client:  clientWith404,
			correct: false,
		},
		{
			name:    "Передаем клиента, который возвращает 404",
			link:    "https://github.com/orlov4919/test",
			client:  clientWith404,
			correct: false,
		},
		{
			name:    "Запрос падает с ошибкой",
			link:    "https://github.com/orlov4919/test",
			client:  clientWithErr,
			correct: false,
		},
		{
			name:    "Получили неверный JSON в ответе",
			link:    "https://github.com/orlov4919/test",
			client:  clientWithWrongJSON,
			correct: false,
		},
		{

			name:    "Получение обновлений выполнено успешно",
			link:    "https://github.com/orlov4919/test",
			client:  clientWithOK,
			correct: true,
			updates: scrapper.LinkUpdates{&scrapper.LinkUpdate{
				CreateTime: "14:39:14 25-02-2025",
				Header:     "Issue",
			}},
		},
	}

	for _, test := range tests {
		gitClient := github.NewClient(testHost, testToken, test.client)
		updates, err := gitClient.LinkUpdates(test.link, time.Now())

		if test.correct {
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.updates, updates)
		} else {
			assert.Error(t, err)
		}
	}
}
