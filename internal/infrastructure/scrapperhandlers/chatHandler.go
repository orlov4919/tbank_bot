package scrapperhandlers

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"linkTraccer/internal/domain/dto"
	"log/slog"
	"net/http"
	"strconv"
)

const (
	contentType      = "Content-Type"
	jsonType         = "application/json"
	wrongID          = "id не соответствует числу"
	negativeID       = "полученное id < 0, должно быть id >=0"
	errId            = "id error"
	idRegistrated    = "id уже зарегистрирован"
	idNotRegistrated = "id не зарегистрирован"
)

type ChatHandler struct {
	userRepo UserRepo
	log      *slog.Logger
}

func NewChatHandler(repo UserRepo, log *slog.Logger) *ChatHandler {
	return &ChatHandler{
		userRepo: repo,
		log:      log,
	}
}

func (c *ChatHandler) HandleChatChanges(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)

	if err != nil {
		c.writeAPIErrToResponse(w, wrongID, err.Error(), http.StatusBadRequest)

		return
	}

	if userID < 0 {
		c.writeAPIErrToResponse(w, errId, negativeID, http.StatusBadRequest)

		return
	}

	userExist, err := c.userRepo.UserExist(userID)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		c.log.Info(
			fmt.Sprintf("обработка запроса %s закончилась ошибкой, при проверке пользователя в БД", r.URL.Path),
			"err", err.Error())

		return
	}

	switch r.Method {
	case http.MethodPost:
		c.PostHandler(w, userID, userExist)

	case http.MethodDelete:
		c.DeleteHandler(w, userID, userExist)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (c *ChatHandler) PostHandler(w http.ResponseWriter, userID int64, userExist bool) {

	if userExist {
		c.writeAPIErrToResponse(w, errId, idRegistrated, http.StatusBadRequest)

		return
	}

	if err := c.userRepo.RegUser(userID); err != nil {
		w.Header().Set(contentType, jsonType)
		w.WriteHeader(http.StatusInternalServerError)

		c.log.Info("ошибка при регистрации пользователя", "err", err.Error())

		return
	}

	w.WriteHeader(http.StatusOK)
}

func (c *ChatHandler) DeleteHandler(w http.ResponseWriter, userID int64, userExist bool) {

	if !userExist {
		c.writeAPIErrToResponse(w, errId, idNotRegistrated, http.StatusNotFound)

		return
	}

	if err := c.userRepo.DeleteUser(userID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		c.log.Info("ошибка при регистрации пользователя", "err", err.Error())

		return
	}

	w.WriteHeader(http.StatusOK)
}

func (c *ChatHandler) writeAPIErrToResponse(w http.ResponseWriter, errName, errDescription string, statusCode int) {
	w.Header().Set(contentType, jsonType)
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(dto.NewAPIErrResponse(errName, errDescription, []string{}))

	if err != nil {
		c.log.Debug("ошибка при формировании JSON APIErrResponse", "err", err.Error())
	}
}
