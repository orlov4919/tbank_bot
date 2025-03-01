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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testHost  = "api.github.com"
	testToken = "123456789"
)

var jsonData = []byte(`[{ "updated_at" : "2025-02-25T11:39:14Z"}]`)
var randomData = []byte("abcdsdfsdf")
var emptyData = []byte(`[]`)
var errTest = errors.New("произошел таймаут")

type Link = scrapper.Link
type LinkState = scrapper.LinkState

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
			name:    "Не корректная ссылка",
			link:    "https://github.com//",
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
	badClient := mocks.NewHTTPClient(t)
	timeoutClient := mocks.NewHTTPClient(t)
	goodClient := mocks.NewHTTPClient(t)

	badClient.On("Do", mock.Anything).Return(&http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(nil)}, nil)
	timeoutClient.On("Do", mock.Anything).Return(nil, errTest)
	goodClient.On("Do", mock.Anything).Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBuffer(jsonData))}, nil)

	type testCase struct {
		name    string
		link    Link
		client  github.HTTPClient
		correct bool
	}

	tests := []testCase{
		{
			name:    "Передаем ссылку с неправильным хостом",
			link:    "https://gitehube.com/orlov4919/test",
			client:  badClient,
			correct: false,
		},
		{
			name:    "Передаем клиента, который таймаутит",
			link:    "https://github.com/orlov4919/test",
			client:  timeoutClient,
			correct: false,
		},
		{
			name:    "Пытаемся отследить репозиторий, которого нет",
			link:    "https://github.com/orlov4919/test1234",
			client:  badClient,
			correct: false,
		},
		{
			name:    "Отслеживаем существующий репозиторий",
			link:    "https://github.com/orlov4919/test",
			client:  goodClient,
			correct: true,
		},
	}

	for _, test := range tests {
		gitClient := github.NewClient(testHost, testToken, test.client)

		assert.Equal(t, test.correct, gitClient.CanTrack(test.link))
	}
}

func TestGitClient_LinkState(t *testing.T) {
	badClient := mocks.NewHTTPClient(t)
	timeoutClient := mocks.NewHTTPClient(t)
	goodClient := mocks.NewHTTPClient(t)
	wrongBodyClient := mocks.NewHTTPClient(t)
	emptyBodyClient := mocks.NewHTTPClient(t)

	badClient.On("Do", mock.Anything).
		Return(&http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(nil)}, nil)
	timeoutClient.On("Do", mock.Anything).
		Return(nil, errTest)
	goodClient.On("Do", mock.Anything).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBuffer(jsonData))}, nil)
	wrongBodyClient.On("Do", mock.Anything).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBuffer(randomData))}, nil)
	emptyBodyClient.On("Do", mock.Anything).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBuffer(emptyData))}, nil)

	type testCase struct {
		name    string
		link    Link
		state   LinkState
		client  github.HTTPClient
		correct bool
	}

	tests := []testCase{
		{
			name:    "Передаем ссылку с неправильным хостом",
			link:    "https://gitehube.com/orlov4919/test",
			client:  badClient,
			correct: false,
		},
		{
			name:    "Передаем клиента, который таймаутит",
			link:    "https://github.com/orlov4919/test",
			client:  timeoutClient,
			correct: false,
		},
		{
			name:    "Пытаемся отследить репозиторий, которого нет",
			link:    "https://github.com/orlov4919/test1234",
			client:  badClient,
			correct: false,
		},
		{
			name:    "Отслеживаем существующий репозиторий",
			link:    "https://github.com/orlov4919/test",
			state:   "2025-02-25T11:39:14Z",
			client:  goodClient,
			correct: true,
		},
		{
			name:    "Сервер прислал неправильный json",
			link:    "https://github.com/orlov4919/test",
			client:  wrongBodyClient,
			correct: false,
		},
		{
			name:    "В репозитории еще нет issue",
			link:    "https://github.com/orlov4919/test",
			client:  emptyBodyClient,
			correct: true,
		},
	}

	for _, test := range tests {
		gitClient := github.NewClient(testHost, testToken, test.client)

		state, err := gitClient.LinkState(test.link)

		assert.Equal(t, test.state, state)

		if test.correct {
			assert.NoError(t, err)
			assert.Equal(t, test.state, state)
		} else {
			assert.Error(t, err)
		}
	}
}
