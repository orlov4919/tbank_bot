package scraphandlers_test

import (
	"bytes"
	"encoding/json"
	"io"
	"linkTraccer/internal/application/scrapper/scrapservice"
	"linkTraccer/internal/domain/dto"
	"linkTraccer/internal/domain/scrapper"
	"linkTraccer/internal/infrastructure/scraphandlers"
	"linkTraccer/internal/infrastructure/scraphandlers/mocks"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	wrongStr   = "Hello Word"
	wrongLink  = "google.com"
	goodLink   = "tbank.com"
	idNotInt   = "hello"
	negativeID = "-5"
	goodID     = "5"
)

var (
	expectedLink = "tbank.ru"

	userLinks = &listLinksResponse{
		Links: []LinkResponse{{
			ID:      0,
			URL:     expectedLink,
			Tags:    []string{},
			Filters: []string{},
		}},
		Size: 1,
	}

	addWrongLink, _     = json.Marshal(scrapper.AddLinkRequest{Link: wrongLink})
	addGoodLink, _      = json.Marshal(scrapper.AddLinkRequest{Link: goodLink})
	addGoodLinkResponse = &scrapper.LinkResponse{
		ID:  1,
		URL: goodLink,
	}

	removeGoodLink, _      = json.Marshal(scrapper.RemoveLinkRequest{Link: goodLink})
	removeGoodLinkResponse = &scrapper.LinkResponse{ID: 1, URL: goodLink}
)

type Link = scrapper.Link
type listLinksResponse = scrapper.ListLinksResponse
type LinkResponse = scrapper.LinkResponse

func TestLinkHandler_GetMethodHandler(t *testing.T) {
	transactor := mocks.NewTransactor(t)

	repoWithErr := mocks.NewUserRepo(t)
	repoWithLinks := mocks.NewUserRepo(t)
	repoWithoutLinks := mocks.NewUserRepo(t)

	stackClient, gitClient := mocks.NewSiteClient(t), mocks.NewSiteClient(t)

	repoWithErr.On("AllUserLinks", mock.Anything).Return(nil, errRepo)
	repoWithLinks.On("AllUserLinks", mock.Anything).Return([]Link{expectedLink}, nil)
	repoWithoutLinks.On("AllUserLinks", mock.Anything).Return([]Link{}, nil)

	type testCase struct {
		name         string
		repo         scrapservice.UserRepo
		userID       int64
		httpStatus   int
		expectedBody *listLinksResponse
	}

	tests := []testCase{
		{
			name:         "ошибка при попытке получить все ссылки пользователя",
			repo:         repoWithErr,
			userID:       1,
			httpStatus:   http.StatusInternalServerError,
			expectedBody: nil,
		},
		{
			name:         "получение ссылок пользователем",
			repo:         repoWithLinks,
			userID:       1,
			httpStatus:   http.StatusOK,
			expectedBody: userLinks,
		},
		{
			name:         "проверка краевого случая, когда ссылок не оказалось",
			repo:         repoWithoutLinks,
			userID:       1,
			httpStatus:   http.StatusOK,
			expectedBody: &listLinksResponse{Links: make([]LinkResponse, 0)},
		},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()

		linkHandler := scraphandlers.NewLinkHandler(test.repo, transactor, logger, stackClient, gitClient)
		linkHandler.GetMethodHandler(w, test.userID)

		assert.Equal(t, test.httpStatus, w.Code)

		if test.expectedBody != nil {
			listLinksResponse := &listLinksResponse{}

			err := json.NewDecoder(w.Body).Decode(listLinksResponse)

			assert.NoError(t, err, "ошибка при декодинге тела ответа")
			assert.Equal(t, test.expectedBody, listLinksResponse)
		} else {
			assert.Empty(t, w.Body)
		}
	}
}

func TestLinkHandler_PostMethodHandler(t *testing.T) {
	transactorWithErr := mocks.NewTransactor(t)
	transactorWithoutErr := mocks.NewTransactor(t)
	//
	transactorWithErr.On("WithTransaction", mock.Anything, mock.Anything).Return(errRepo)
	transactorWithoutErr.On("WithTransaction", mock.Anything, mock.Anything).Return(nil)

	repoWithErrUserTrackLink := mocks.NewUserRepo(t)
	repoUserAlwaysTrackLink := mocks.NewUserRepo(t)
	repoUserAlwaysNotTrackLink := mocks.NewUserRepo(t)

	repoWithErrUserTrackLink.On("UserTrackLink", mock.Anything, mock.Anything).
		Return(false, errRepo)

	repoUserAlwaysTrackLink.On("UserTrackLink", mock.Anything, mock.Anything).
		Return(true, nil)

	repoUserAlwaysNotTrackLink.On("UserTrackLink", mock.Anything, mock.Anything).
		Return(false, nil)

	stackClient := mocks.NewSiteClient(t)

	stackClient.On("CanTrack", wrongLink).Return(false)
	stackClient.On("CanTrack", goodLink).Return(true)

	type testCase struct {
		name           string
		userRepo       scrapservice.UserRepo
		transactor     scrapservice.Transactor
		userID         int64
		httpStatus     int
		reqData        []byte
		responseAPIErr bool
		responseLink   bool
		expectedBody   any
	}

	tests := []testCase{
		{
			name:           "Передаем не правильные данные в запросе",
			userRepo:       repoWithErrUserTrackLink,
			userID:         1,
			httpStatus:     http.StatusBadRequest,
			reqData:        []byte(wrongStr),
			responseAPIErr: true,
			responseLink:   false,
			expectedBody:   dto.ApiErrBadJSON,
		},
		{
			name:           "пытаемся добавить не поддерживаемую ссылку",
			userRepo:       repoWithErrUserTrackLink,
			userID:         1,
			httpStatus:     http.StatusBadRequest,
			reqData:        addWrongLink,
			responseAPIErr: true,
			responseLink:   false,
			expectedBody:   dto.ApiErrBadLink,
		},
		{
			name:           "ошибка при проверке отслеживания ссылки",
			userRepo:       repoWithErrUserTrackLink,
			userID:         1,
			httpStatus:     http.StatusInternalServerError,
			reqData:        addGoodLink,
			responseAPIErr: false,
			responseLink:   false,
			expectedBody:   nil,
		},
		{
			name:           "пользователь уже отслеживает ссылку",
			userRepo:       repoUserAlwaysTrackLink,
			userID:         1,
			httpStatus:     http.StatusBadRequest,
			reqData:        addGoodLink,
			responseAPIErr: true,
			responseLink:   false,
			expectedBody:   dto.ApiErrDuplicateLink,
		},
		{
			name:           "ошибка при добавлении новой ссылки в БД",
			userRepo:       repoUserAlwaysNotTrackLink,
			transactor:     transactorWithErr,
			userID:         1,
			httpStatus:     http.StatusInternalServerError,
			reqData:        addGoodLink,
			responseAPIErr: false,
			responseLink:   false,
			expectedBody:   nil,
		},
		{
			name:           "успешное добавлении новой ссылки в БД",
			transactor:     transactorWithoutErr,
			userRepo:       repoUserAlwaysNotTrackLink,
			userID:         1,
			httpStatus:     http.StatusOK,
			reqData:        addGoodLink,
			responseAPIErr: false,
			responseLink:   true,
			expectedBody:   addGoodLinkResponse,
		},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()

		linkHandler := scraphandlers.NewLinkHandler(test.userRepo, test.transactor, logger, stackClient)
		linkHandler.PostMethodHandler(w, test.userID, test.reqData)

		assert.Equal(t, test.httpStatus, w.Code)

		switch {
		case test.responseAPIErr:
			errResponse := &dto.APIErrResponse{}

			err := json.NewDecoder(w.Body).Decode(errResponse)

			assert.NoError(t, err)
			assert.Equal(t, test.expectedBody, errResponse)
		case test.responseLink:
			linksResponse := &LinkResponse{}

			err := json.NewDecoder(w.Body).Decode(linksResponse)

			assert.NoError(t, err)
			assert.Equal(t, test.expectedBody, linksResponse)
		default:
			assert.Empty(t, w.Body)
		}
	}
}

func TestLinkHandler_DeleteMethodHandler(t *testing.T) {
	transactor := mocks.NewTransactor(t)

	repoWithErrUserTrackLink := mocks.NewUserRepo(t)
	repoUserAlwaysNotTrackLink := mocks.NewUserRepo(t)
	repoUntrackLinkWithErr := mocks.NewUserRepo(t)
	repoUntrackLink := mocks.NewUserRepo(t)

	repoWithErrUserTrackLink.On("UserTrackLink", mock.Anything, mock.Anything).
		Return(false, errRepo)

	repoUserAlwaysNotTrackLink.On("UserTrackLink", mock.Anything, mock.Anything).
		Return(false, nil)

	repoUntrackLinkWithErr.On("UserTrackLink", mock.Anything, mock.Anything).
		Return(true, nil)

	repoUntrackLinkWithErr.On("UntrackLink", mock.Anything, mock.Anything).
		Return(errRepo)

	repoUntrackLink.On("UserTrackLink", mock.Anything, mock.Anything).
		Return(true, nil)

	repoUntrackLink.On("UntrackLink", mock.Anything, mock.Anything).
		Return(nil)

	stackClient := mocks.NewSiteClient(t)

	type testCase struct {
		name           string
		userRepo       scrapservice.UserRepo
		userID         int64
		httpStatus     int
		reqData        []byte
		responseAPIErr bool
		responseLink   bool
		expectedBody   any
	}

	tests := []testCase{
		{
			name:           "Передаем не правильные данные в запросе",
			userRepo:       repoWithErrUserTrackLink,
			userID:         1,
			httpStatus:     http.StatusBadRequest,
			reqData:        []byte(wrongStr),
			responseAPIErr: true,
			responseLink:   false,
			expectedBody:   dto.ApiErrBadJSON,
		},
		{
			name:           "ошибка при проверке отслеживания ссылки",
			userRepo:       repoWithErrUserTrackLink,
			userID:         1,
			httpStatus:     http.StatusInternalServerError,
			reqData:        removeGoodLink,
			responseAPIErr: false,
			responseLink:   false,
			expectedBody:   nil,
		},
		{
			name:           "пользователь не отслеживает ссылку",
			userRepo:       repoUserAlwaysNotTrackLink,
			userID:         1,
			httpStatus:     http.StatusNotFound,
			reqData:        removeGoodLink,
			responseAPIErr: true,
			responseLink:   false,
			expectedBody:   dto.ApiErrNotTrackLink,
		},
		{
			name:           "ошибка при удалении ссылки пользователя",
			userRepo:       repoUntrackLinkWithErr,
			userID:         1,
			httpStatus:     http.StatusInternalServerError,
			reqData:        removeGoodLink,
			responseAPIErr: false,
			responseLink:   false,
			expectedBody:   nil,
		},
		{
			name:           "успешное удаление ссылки",
			userRepo:       repoUntrackLink,
			userID:         1,
			httpStatus:     http.StatusOK,
			reqData:        removeGoodLink,
			responseAPIErr: false,
			responseLink:   true,
			expectedBody:   removeGoodLinkResponse,
		},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()

		linkHandler := scraphandlers.NewLinkHandler(test.userRepo, transactor, logger, stackClient)
		linkHandler.DeleteMethodHandler(w, test.userID, test.reqData)

		assert.Equal(t, test.httpStatus, w.Code)

		switch {
		case test.responseAPIErr:
			errResponse := &dto.APIErrResponse{}

			err := json.NewDecoder(w.Body).Decode(errResponse)

			assert.NoError(t, err)
			assert.Equal(t, test.expectedBody, errResponse)
		case test.responseLink:
			linksResponse := &LinkResponse{}

			err := json.NewDecoder(w.Body).Decode(linksResponse)

			assert.NoError(t, err)
			assert.Equal(t, test.expectedBody, linksResponse)

		default:
			assert.Empty(t, w.Body)
		}
	}
}

func TestLinkHandler_HandleLinksChanges(t *testing.T) {
	transactor := mocks.NewTransactor(t)

	userRepoWithErr := mocks.NewUserRepo(t)
	userRepoWithoutUsers := mocks.NewUserRepo(t)
	userRepoWithUsers := mocks.NewUserRepo(t)

	userRepoWithErr.On("UserExist", mock.Anything).Return(false, errRepo)
	userRepoWithoutUsers.On("UserExist", mock.Anything).Return(false, nil)
	userRepoWithUsers.On("UserExist", mock.Anything).Return(true, nil)
	userRepoWithUsers.On("AllUserLinks", mock.Anything).Return(nil, errRepo)

	stackClient := mocks.NewSiteClient(t)

	type testCase struct {
		name         string
		userRepo     scrapservice.UserRepo
		req          *http.Request
		httpStatus   int
		responseBody any
	}

	tests := []testCase{
		{
			name:     "id пользователя не является числом",
			userRepo: userRepoWithErr,
			req: &http.Request{
				Body:   io.NopCloser(bytes.NewBuffer([]byte{})),
				Header: map[string][]string{"Tg-Chat-Id": {idNotInt}},
			},
			httpStatus:   http.StatusBadRequest,
			responseBody: dto.ApiErrIDNotNum,
		},
		{
			name:     "id пользователя отрицательное число",
			userRepo: userRepoWithErr,
			req: &http.Request{
				Body:   io.NopCloser(bytes.NewBuffer([]byte{})),
				Header: map[string][]string{"Tg-Chat-Id": {negativeID}},
			},
			httpStatus:   http.StatusBadRequest,
			responseBody: dto.ApiErrNegativeID,
		},
		{
			name:     "id пользователя отрицательное число",
			userRepo: userRepoWithErr,
			req: &http.Request{
				Body:   io.NopCloser(bytes.NewBuffer([]byte{})),
				Header: map[string][]string{"Tg-Chat-Id": {goodID}},
			},
			httpStatus:   http.StatusInternalServerError,
			responseBody: nil,
		},
		{
			name:     "id пользователя не зарегистрирован",
			userRepo: userRepoWithoutUsers,
			req: &http.Request{
				Body:   io.NopCloser(bytes.NewBuffer([]byte{})),
				Header: map[string][]string{"Tg-Chat-Id": {goodID}},
			},
			httpStatus:   http.StatusBadRequest,
			responseBody: dto.ApiErrUserNotRegistered,
		},
		{
			name:     "проверка не поддерживаемого метода",
			userRepo: userRepoWithUsers,
			req: &http.Request{
				Body:   io.NopCloser(bytes.NewBuffer([]byte{})),
				Header: map[string][]string{"Tg-Chat-Id": {goodID}},
				Method: http.MethodPut,
			},
			httpStatus:   http.StatusMethodNotAllowed,
			responseBody: nil,
		},
		{
			name:     "проверка вызова get обработчика",
			userRepo: userRepoWithUsers,
			req: &http.Request{
				Body:   io.NopCloser(bytes.NewBuffer([]byte{})),
				Header: map[string][]string{"Tg-Chat-Id": {goodID}},
				Method: http.MethodGet,
			},
			httpStatus:   http.StatusInternalServerError,
			responseBody: nil,
		},
		{
			name:     "проверка вызова post обработчика",
			userRepo: userRepoWithUsers,
			req: &http.Request{
				Body:   io.NopCloser(bytes.NewBuffer([]byte(wrongStr))),
				Header: map[string][]string{"Tg-Chat-Id": {goodID}},
				Method: http.MethodPost,
			},
			httpStatus:   http.StatusBadRequest,
			responseBody: dto.ApiErrBadJSON,
		},
		{
			name:     "проверка вызова delete обработчика",
			userRepo: userRepoWithUsers,
			req: &http.Request{
				Body:   io.NopCloser(bytes.NewBuffer([]byte(wrongStr))),
				Header: map[string][]string{"Tg-Chat-Id": {goodID}},
				Method: http.MethodPost,
			},
			httpStatus:   http.StatusBadRequest,
			responseBody: dto.ApiErrBadJSON,
		},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()

		linkHandler := scraphandlers.NewLinkHandler(test.userRepo, transactor, logger, stackClient)

		linkHandler.HandleLinksChanges(w, test.req)

		assert.Equal(t, test.httpStatus, w.Code)

		if test.responseBody != nil {
			apiErr := &dto.APIErrResponse{}

			err := json.NewDecoder(w.Body).Decode(apiErr)

			assert.NoError(t, err)
			assert.Equal(t, test.responseBody, apiErr)
		} else {
			assert.Empty(t, w.Body)
		}
	}
}
