package bothandler_test

import (
	"bytes"
	"encoding/json"
	"io"
	"linkTraccer/internal/domain/dto"
	"linkTraccer/internal/infrastructure/bothandler"
	"linkTraccer/internal/infrastructure/bothandler/mocks"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	logLevel = slog.LevelDebug
	logger   = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))

	wrongJSON = []byte("Hello World")

	linkUpdateJSON, _ = json.Marshal(&dto.LinkUpdate{})
)

func TestUpdateServer_HandleLinkUpdates(t *testing.T) {
	tgClient := mocks.NewTgClient(t)
	botHandler := bothandler.New(tgClient, logger)

	type testCase struct {
		name         string
		r            *http.Request
		responseBody *dto.APIErrResponse
		httpStatus   int
	}

	tests := []testCase{
		{
			name:         "некорректный json в запросе",
			r:            &http.Request{Method: http.MethodPost, Body: io.NopCloser(bytes.NewBuffer(wrongJSON))},
			responseBody: dto.APIErrBadJSON,
			httpStatus:   http.StatusBadRequest,
		},
		{
			name:         "передаем валидные данные",
			r:            &http.Request{Method: http.MethodPost, Body: io.NopCloser(bytes.NewBuffer(linkUpdateJSON))},
			responseBody: nil,
			httpStatus:   http.StatusOK,
		},
		{
			name:         "отправляем не обрабатываемый запрос",
			r:            &http.Request{Method: http.MethodGet, Body: io.NopCloser(bytes.NewBuffer(linkUpdateJSON))},
			responseBody: nil,
			httpStatus:   http.StatusMethodNotAllowed,
		},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()

		botHandler.HandleLinkUpdates(w, test.r)

		assert.Equal(t, test.httpStatus, w.Code)

		if test.responseBody != nil {
			respBody := &dto.APIErrResponse{}

			err := json.NewDecoder(w.Body).Decode(respBody)

			assert.NoError(t, err)
			assert.Equal(t, test.responseBody, respBody)
		} else {
			assert.Empty(t, w.Body)
		}
	}
}
