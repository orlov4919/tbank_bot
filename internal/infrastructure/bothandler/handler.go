package bothandler

import (
	"encoding/json"
	"fmt"
	"io"
	"linkTraccer/internal/application/botService"
	"linkTraccer/internal/domain/dto"
	"log"
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
	tgClient botService.TgClient
}

func New(client botService.TgClient) *UpdatesHandler {
	return &UpdatesHandler{tgClient: client}
}

func (s *UpdatesHandler) HandleLinkUpdates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse := dto.NewApiErrResponse(methodError, fmt.Sprintf(methodErrorDescription, r.Method), []string{})

		WriteInResponse(w, http.StatusMethodNotAllowed, errorResponse)

		return
	}

	bodyData, err := io.ReadAll(r.Body)

	if err != nil {
		errorResponse := dto.NewApiErrResponse(ReadBodyError, err.Error(), []string{})

		WriteInResponse(w, http.StatusBadRequest, errorResponse)

		return
	}

	defer r.Body.Close()

	linkUpdate := &dto.LinkUpdate{}

	if err = json.Unmarshal(bodyData, linkUpdate); err != nil {
		errorResponse := dto.NewApiErrResponse(jsonError, err.Error(), []string{})

		WriteInResponse(w, http.StatusBadRequest, errorResponse)
	} else {
		WriteInResponse(w, http.StatusOK, nil)

		for _, userID := range linkUpdate.TgChatIds { // переписать на горутины
			err := s.tgClient.SendMessage(userID, linkUpdate.Description)

			if err != nil {
				log.Println("Ошибка при отправке обновлений пользователю")
			}
		}
	}
}

func WriteInResponse(w http.ResponseWriter, httpStatus int, data any) {
	w.WriteHeader(httpStatus)

	if httpStatus == http.StatusOK {
		return
	}

	if data != nil {
		w.Header().Set(contentType, jsonType)

		err := json.NewEncoder(w).Encode(data)

		if err != nil {
			log.Println(fmt.Errorf("при формировании json ответа произошла ошибка: %w", err))
		}
	}
}
