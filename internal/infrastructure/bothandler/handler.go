package bothandler

import (
	"encoding/json"
	"fmt"
	"io"
	"linkTraccer/internal/application/botservice"
	"linkTraccer/internal/domain/dto"
	"log/slog"
	"net/http"
)

const (
	ReadBodyError          = "ошибка при чтении тела запроса"
	contentType            = "Content-Type"
	jsonType               = "application/json"
	jsonError              = "json, который пришел в запросе не соответствует описанию LinkUpdate"
	methodError            = "нет доступа к запрашиваемому методу"
	methodErrorDescription = `Вы пытаетесь вызвать метод %s, который не поддерживает данный endpoint.
                              Поддерживаемые методы: POST`
)

type UpdatesHandler struct {
	tgClient botservice.TgClient
	log      *slog.Logger
}

func New(client botservice.TgClient, log *slog.Logger) *UpdatesHandler {
	return &UpdatesHandler{
		tgClient: client,
		log:      log,
	}
}

func (s *UpdatesHandler) HandleLinkUpdates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse := dto.NewAPIErrResponse(methodError, fmt.Sprintf(methodErrorDescription, r.Method), []string{})

		s.WriteInResponse(w, http.StatusMethodNotAllowed, errorResponse)

		return
	}

	bodyData, err := io.ReadAll(r.Body)

	if err != nil {
		errorResponse := dto.NewAPIErrResponse(ReadBodyError, err.Error(), []string{})

		s.WriteInResponse(w, http.StatusBadRequest, errorResponse)

		return
	}

	defer r.Body.Close()

	linkUpdate := &dto.LinkUpdate{}

	if err = json.Unmarshal(bodyData, linkUpdate); err != nil {
		errorResponse := dto.NewAPIErrResponse(jsonError, err.Error(), []string{})

		s.WriteInResponse(w, http.StatusBadRequest, errorResponse)
	} else {
		s.WriteInResponse(w, http.StatusOK, nil)

		for _, userID := range linkUpdate.TgChatIDs { // переписать на горутины
			err := s.tgClient.SendMessage(userID, linkUpdate.Description+": "+linkUpdate.URL)

			if err != nil {
				s.log.Debug("ошибка при отправке обновлений пользователю", "err", err.Error())
			}
		}
	}
}

func (s *UpdatesHandler) WriteInResponse(w http.ResponseWriter, httpStatus int, data any) {
	w.WriteHeader(httpStatus)

	if httpStatus == http.StatusOK {
		return
	}

	if data != nil {
		w.Header().Set(contentType, jsonType)

		err := json.NewEncoder(w).Encode(data)

		if err != nil {
			s.log.Debug("при формировании json ответа произошла ошибка", "err", err.Error())
		}
	}
}
