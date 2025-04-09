package bothandler

import (
	"encoding/json"
	"io"
	"linkTraccer/internal/application/botservice"
	"linkTraccer/internal/domain/dto"
	"log/slog"
	"net/http"
)

const (
	contentType = "Content-Type"
	jsonType    = "application/json"
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

func (u *UpdatesHandler) HandleLinkUpdates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)

		return
	}

	bodyData, err := io.ReadAll(r.Body)

	if err != nil {
		u.APIErrToResponse(w, dto.APIErrCantReadBody, http.StatusBadRequest)

		return
	}

	defer r.Body.Close()

	linkUpdate := &dto.LinkUpdate{}

	if err = json.Unmarshal(bodyData, linkUpdate); err != nil {
		u.APIErrToResponse(w, dto.APIErrBadJSON, http.StatusBadRequest)

		return
	}

	w.WriteHeader(http.StatusOK)

	for _, userID := range linkUpdate.TgChatIDs { // переписать на горутины
		err := u.tgClient.SendMessage(userID, linkUpdate.Description+linkUpdate.URL)

		if err != nil {
			u.log.Error("ошибка при отправке обновлений по ссылке в телеграмм", "err", err.Error())
		}
	}
}

func (u *UpdatesHandler) APIErrToResponse(w http.ResponseWriter, errAPI *dto.APIErrResponse, statusCode int) {
	w.Header().Set(contentType, jsonType)
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(errAPI)

	if err != nil {
		u.log.Error("ошибка при формировании JSON APIErrResponse", "err", err.Error())
	}
}
