package scraphandlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"linkTraccer/internal/application/scrapper/scrapservice"
	"linkTraccer/internal/domain/dto"
	"linkTraccer/internal/infrastructure/scraphandlers"
	"linkTraccer/internal/infrastructure/scraphandlers/mocks"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	logLevel = slog.LevelDebug

	logger  = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	errRepo = errors.New("ошибка в репозитории")
)

func TestChatHandler_PostHandler(t *testing.T) {
	transactor := mocks.NewTransactor(t)

	repoWithErr := mocks.NewUserRepo(t)
	repoWithoutErr := mocks.NewUserRepo(t)

	repoWithErr.On("RegUser", mock.Anything).Return(errRepo)
	repoWithoutErr.On("RegUser", mock.Anything).Return(nil)

	type TestCase struct {
		name           string
		userID         int64
		userExist      bool
		repo           scrapservice.UserRepo
		expectedBody   *dto.APIErrResponse
		expectedStatus int
	}

	tests := []TestCase{
		{
			name:           "пытаемся зарегистрировать, зарегистрированного пользователя",
			userID:         1,
			userExist:      true,
			repo:           nil,
			expectedBody:   dto.APIErrUserRegistered,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "ошибка при регистрации пользователя",
			userID:         2,
			userExist:      false,
			repo:           repoWithErr,
			expectedBody:   nil,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "пользователь успешно сохранен",
			userID:         2,
			userExist:      false,
			repo:           repoWithoutErr,
			expectedBody:   nil,
			expectedStatus: http.StatusOK,
		},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()

		chatHandler := scraphandlers.NewChatHandler(test.repo, transactor, logger)

		chatHandler.PostHandler(w, test.userID, test.userExist)

		assert.Equal(t, test.expectedStatus, w.Code)

		if test.expectedBody != nil {
			unmarshalBody := &dto.APIErrResponse{}

			err := json.Unmarshal(w.Body.Bytes(), unmarshalBody)

			assert.NoError(t, err, "ошибка при анмаршалинге тела ответа")
			assert.Equal(t, test.expectedBody, unmarshalBody)
		} else {
			assert.Empty(t, w.Body.String())
		}
	}
}

func TestChatHandler_DeleteHandler(t *testing.T) {
	transactorWithoutErr := mocks.NewTransactor(t)
	transactorWithErr := mocks.NewTransactor(t)

	transactorWithoutErr.On("WithTransaction", mock.Anything, mock.Anything).Return(nil)
	transactorWithErr.On("WithTransaction", mock.Anything, mock.Anything).Return(errRepo)

	repoWithoutErr := mocks.NewUserRepo(t)

	type TestCase struct {
		name           string
		userID         int64
		userExist      bool
		repo           scrapservice.UserRepo
		transactor     scrapservice.Transactor
		expectedBody   *dto.APIErrResponse
		expectedStatus int
	}

	tests := []TestCase{
		{
			name:           "пытаемся удалить, не зарегистрированного пользователя",
			userID:         1,
			userExist:      false,
			repo:           nil,
			expectedBody:   dto.APIErrUserNotRegistered,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "ошибка при удалении пользователя из БД",
			userID:         2,
			userExist:      true,
			transactor:     transactorWithErr,
			repo:           repoWithoutErr,
			expectedBody:   nil,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "пользователь успешно удален",
			userID:         2,
			userExist:      true,
			repo:           repoWithoutErr,
			transactor:     transactorWithoutErr,
			expectedBody:   nil,
			expectedStatus: http.StatusOK,
		},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()

		chatHandler := scraphandlers.NewChatHandler(test.repo, test.transactor, logger)

		chatHandler.DeleteHandler(w, test.userID, test.userExist)

		assert.Equal(t, test.expectedStatus, w.Code)

		if test.expectedBody != nil {
			unmarshalBody := &dto.APIErrResponse{}

			err := json.Unmarshal(w.Body.Bytes(), unmarshalBody)

			assert.NoError(t, err, "ошибка при анмаршалинге тела ответа")
			assert.Equal(t, test.expectedBody, unmarshalBody)
		} else {
			assert.Empty(t, w.Body.String())
		}
	}
}

func TestChatHandler_HandleChatChanges(t *testing.T) {
	transactor := mocks.NewTransactor(t)

	transactor.On("WithTransaction", mock.Anything, mock.Anything).Return(nil)

	repoWithErr := mocks.NewUserRepo(t)
	repoWithoutUsers := mocks.NewUserRepo(t)
	repoWithUsers := mocks.NewUserRepo(t)

	repoWithErr.On("UserExist", mock.Anything).Return(false, errRepo)
	repoWithoutUsers.On("UserExist", mock.Anything).Return(false, nil)
	repoWithoutUsers.On("RegUser", mock.Anything).Return(nil)
	repoWithUsers.On("UserExist", mock.Anything).Return(true, nil)

	type TestCase struct {
		name           string
		userID         string
		repo           scrapservice.UserRepo
		httpMethod     string
		expectedStatus int
		expectedBody   *dto.APIErrResponse
	}

	tests := []TestCase{
		{
			name:           "пришло не числовое id",
			userID:         "Hello Word",
			repo:           nil,
			httpMethod:     http.MethodPost,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   dto.APIErrIDNotNum,
		},
		{
			name:           "передаем отрицательное id",
			userID:         "-5",
			repo:           nil,
			httpMethod:     http.MethodPost,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   dto.APIErrNegativeID,
		},
		{
			name:           "ошибка в БД при проверке пользователя ",
			userID:         "1",
			repo:           repoWithErr,
			httpMethod:     http.MethodPost,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   nil,
		},
		{
			name:           "обрабатываем запрос на добавление, еще не добавленного пользователя",
			userID:         "5",
			repo:           repoWithoutUsers,
			httpMethod:     http.MethodPost,
			expectedStatus: http.StatusOK,
			expectedBody:   nil,
		},
		{
			name:           "обрабатываем запрос на удаление пользователя, который уже добавлен",
			userID:         "10",
			repo:           repoWithUsers,
			httpMethod:     http.MethodDelete,
			expectedStatus: http.StatusOK,
			expectedBody:   nil,
		},
		{
			name:           "обрабатываем метод, который не поддерживается",
			userID:         "10",
			repo:           repoWithUsers,
			httpMethod:     http.MethodPut,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   nil,
		},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(test.httpMethod, "", io.NopCloser(new(bytes.Buffer)))

		vars := map[string]string{
			"id": test.userID,
		}

		r = mux.SetURLVars(r, vars)

		chatHandler := scraphandlers.NewChatHandler(test.repo, transactor, logger)

		chatHandler.HandleChatChanges(w, r)

		assert.Equal(t, test.expectedStatus, w.Code)

		if test.expectedBody != nil {
			unmarshalBody := &dto.APIErrResponse{}

			err := json.Unmarshal(w.Body.Bytes(), unmarshalBody)

			assert.NoError(t, err, "ошибка при анмаршалинге тела ответа")
			assert.Equal(t, test.expectedBody, unmarshalBody)
		} else {
			assert.Empty(t, w.Body.String())
		}
	}
}
