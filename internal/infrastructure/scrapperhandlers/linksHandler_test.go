package scrapperhandlers_test

import (
	"bytes"
	"encoding/json"
	"io"
	"linkTraccer/internal/domain/scrapper"
	"linkTraccer/internal/infrastructure/database/file/userstorage"
	"linkTraccer/internal/infrastructure/scrapperhandlers"
	"linkTraccer/internal/infrastructure/scrapperhandlers/mocks"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	regID          = 3
	errID          = 2
	goodID         = 1
	errIDStr       = "2"
	goodIDStr      = "1"
	expectedLink   = "google.com"
	unexpectedLink = "tbank.ru"
	wrongStr       = "wrong data"
	newLink        = "stackoverflov.com"
)

var trackedLink, _ = json.Marshal(addLinkRequest{Link: "google.com"})
var notSupportLink, _ = json.Marshal(addLinkRequest{Link: "tbank.ru"})
var newLinkJSON, _ = json.Marshal(addLinkRequest{Link: "stackoverflov.com"})
var removeAddedLink, _ = json.Marshal(removeLinkRequest{Link: "google.com"})
var removeNotAddedLink, _ = json.Marshal(removeLinkRequest{Link: "tbank.ru"})

type link = scrapper.Link
type listLinksResponse = scrapper.ListLinksResponse
type linkResponse = scrapper.LinkResponse
type addLinkRequest = scrapper.AddLinkRequest
type removeLinkRequest = scrapper.RemoveLinkRequest

func TestLinkHandler_GetMethodHandler(t *testing.T) {
	userRepo := mocks.NewUserRepo(t)
	stackClient, gitClient := mocks.NewSiteClient(t), mocks.NewSiteClient(t)

	userRepo.On("AllUserLinks", errID).Return(nil, userstorage.NewErrWithStorage("нет id"))
	userRepo.On("AllUserLinks", goodID).Return([]link{expectedLink}, nil)
	userRepo.On("AllUserLinks", regID).Return([]link{}, nil)

	linkHandler := scrapperhandlers.NewLinkHandler(userRepo, logger, stackClient, gitClient)

	type testCase struct {
		name       string
		userID     int
		httpStatus int
		correct    bool
		answer     listLinksResponse
	}

	tests := []testCase{
		{
			name:       "Пытаемся получить ссылки незарегистрированного пользователя",
			userID:     errID,
			httpStatus: http.StatusBadRequest,
			correct:    false,
		},
		{
			name:       "Получаем ссылки зарегистрированного пользователя c сохраненными ссылками",
			userID:     goodID,
			httpStatus: http.StatusOK,
			correct:    true,
			answer: listLinksResponse{
				Links: []linkResponse{{
					ID:      0,
					URL:     expectedLink,
					Tags:    []string{},
					Filters: []string{},
				}},
				Size: 1,
			},
		},
		{
			name:       "Получаем ссылки зарегистрированного пользователя без ссылок",
			userID:     regID,
			httpStatus: http.StatusOK,
			correct:    true,
			answer: listLinksResponse{
				Links: []linkResponse{},
				Size:  0,
			},
		},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()

		linkHandler.GetMethodHandler(w, test.userID)
		assert.Equal(t, test.httpStatus, w.Code)

		if test.correct {
			listLinksResponse := &listLinksResponse{}

			_ = json.NewDecoder(w.Body).Decode(listLinksResponse)

			assert.Equal(t, test.answer.Size, listLinksResponse.Size)
			assert.ElementsMatch(t, test.answer.Links, listLinksResponse.Links)
		}
	}
}

func TestLinkHandler_PostMethodHandler(t *testing.T) {
	userRepo := mocks.NewUserRepo(t)

	stackClient := mocks.NewSiteClient(t)

	stackClient.On("CanTrack", unexpectedLink).Return(false)
	stackClient.On("CanTrack", newLink).Return(true)

	userRepo.On("UserTrackLink", goodID, expectedLink).Return(true)
	userRepo.On("UserTrackLink", goodID, unexpectedLink).Return(false)
	userRepo.On("UserTrackLink", goodID, newLink).Return(false)
	userRepo.On("TrackLink", goodID, newLink, mock.Anything).Return(nil)

	linkHandler := scrapperhandlers.NewLinkHandler(userRepo, logger, stackClient)

	type testCase struct {
		name       string
		userID     int
		httpStatus int
		bodyData   []byte
		correct    bool
		answer     linkResponse
	}

	tests := []testCase{
		{
			name:       "Передаем не правильные данные в запросе",
			userID:     goodID,
			httpStatus: http.StatusBadRequest,
			bodyData:   []byte(wrongStr),
			correct:    false,
		},
		{
			name:       "Передаем ссылку, которую пользователь уже отслеживает",
			userID:     goodID,
			httpStatus: http.StatusBadRequest,
			bodyData:   trackedLink,
			correct:    false,
		},
		{
			name:       "Передаем ссылку, которая не поддерживается клиентами",
			userID:     goodID,
			httpStatus: http.StatusBadRequest,
			bodyData:   notSupportLink,
			correct:    false,
		},
		{
			name:       "Передаем ссылку, которая поддерживается и которую еще не отслеживает",
			userID:     goodID,
			httpStatus: http.StatusOK,
			bodyData:   newLinkJSON,
			answer: linkResponse{
				ID:  1,
				URL: newLink,
			},
			correct: true,
		},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()

		linkHandler.PostMethodHandler(w, test.userID, test.bodyData)
		assert.Equal(t, test.httpStatus, w.Code)

		if test.correct {
			linksResponse := &linkResponse{}

			_ = json.NewDecoder(w.Body).Decode(linksResponse)

			assert.Equal(t, test.answer, *linksResponse)
		}
	}
}

func TestLinkHandler_DeleteMethodHandler(t *testing.T) {
	userRepo := mocks.NewUserRepo(t)
	stackClient := mocks.NewSiteClient(t)

	userRepo.On("UserTrackLink", goodID, expectedLink).Return(true)
	userRepo.On("UserTrackLink", goodID, unexpectedLink).Return(false)
	userRepo.On("UntrackLink", goodID, expectedLink).Return(nil)

	linkHandler := scrapperhandlers.NewLinkHandler(userRepo, logger, stackClient)

	type testCase struct {
		name       string
		userID     int
		httpStatus int
		bodyData   []byte
		correct    bool
		answer     linkResponse
	}

	tests := []testCase{
		{
			name:       "Передаем данные не соответствующие json",
			userID:     goodID,
			httpStatus: http.StatusBadRequest,
			bodyData:   []byte(wrongStr),
			correct:    false,
		},
		{
			name:       "Пытаемся удалить не добавленную ссылку",
			userID:     goodID,
			httpStatus: http.StatusNotFound,
			bodyData:   removeNotAddedLink,
			correct:    false,
		},
		{
			name:       "Пытаемся удалить добавленную ссылку",
			userID:     goodID,
			httpStatus: http.StatusOK,
			bodyData:   removeAddedLink,
			correct:    false,
			answer: linkResponse{
				ID:  1,
				URL: expectedLink,
			},
		},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()

		linkHandler.DeleteMethodHandler(w, test.userID, test.bodyData)
		assert.Equal(t, test.httpStatus, w.Code)

		if test.correct {
			linksResponse := &linkResponse{}

			_ = json.NewDecoder(w.Body).Decode(linksResponse)

			assert.Equal(t, test.answer, *linksResponse)
		}
	}
}

func TestLinkHandler_HandleLinksChanges(t *testing.T) {
	userRepo := mocks.NewUserRepo(t)
	stackClient := mocks.NewSiteClient(t)

	userRepo.On("AllUserLinks", errID).Return(nil, userstorage.NewErrWithStorage("нет ссылок"))

	linkHandler := scrapperhandlers.NewLinkHandler(userRepo, logger, stackClient)

	type testCase struct {
		name       string
		req        *http.Request
		httpStatus int
	}

	tests := []testCase{
		{
			name:       "Передаем запрос с методом, который не поддерживается",
			req:        &http.Request{Method: http.MethodPut},
			httpStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "Передаем запрос без айди",
			req:        &http.Request{Method: http.MethodGet, Body: io.NopCloser(bytes.NewBuffer([]byte{}))},
			httpStatus: http.StatusBadRequest,
		},
		{
			name: "Передаем Get запрос, который вернет ошибку",
			req: &http.Request{
				Method: http.MethodGet,
				Body:   io.NopCloser(bytes.NewBuffer([]byte{})),
				Header: map[string][]string{"Tg-Chat-Id": {errIDStr}},
			},
			httpStatus: http.StatusBadRequest,
		},
		{
			name: "Передаем Post запрос, который вернет ошибку",
			req: &http.Request{
				Method: http.MethodPost,
				Body:   io.NopCloser(bytes.NewBuffer([]byte{})),
				Header: map[string][]string{"Tg-Chat-Id": {errIDStr}},
			},
			httpStatus: http.StatusBadRequest,
		},
		{
			name: "Передаем Delete запрос, который вернет ошибку",
			req: &http.Request{Method: http.MethodDelete,
				Body:   io.NopCloser(bytes.NewBuffer([]byte{})),
				Header: map[string][]string{"Tg-Chat-Id": {errIDStr}}},
			httpStatus: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()
		linkHandler.HandleLinksChanges(w, test.req)

		assert.Equal(t, test.httpStatus, w.Code)
	}
}
