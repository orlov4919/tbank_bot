package scraphandlers

import (
	"context"
	"encoding/json"
	"fmt"
	"linkTraccer/internal/domain/dto"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

const (
	contentType = "Content-Type"
	jsonType    = "application/json"
)

type ChatHandler struct {
	userRepo   UserRepo
	transactor Transactor
	log        *slog.Logger
}

func NewChatHandler(repo UserRepo, transactor Transactor, log *slog.Logger) *ChatHandler {
	return &ChatHandler{
		userRepo:   repo,
		transactor: transactor,
		log:        log,
	}
}

func (c *ChatHandler) HandleChatChanges(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)

	defer r.Body.Close()

	if err != nil {
		c.apiErrToResponse(w, dto.APIErrIDNotNum, http.StatusBadRequest)

		return
	}

	if userID < 0 {
		c.apiErrToResponse(w, dto.APIErrNegativeID, http.StatusBadRequest)

		return
	}

	userExist, err := c.userRepo.UserExist(userID)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		c.log.Error(
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
		c.apiErrToResponse(w, dto.APIErrUserRegistered, http.StatusBadRequest)

		return
	}

	if err := c.userRepo.RegUser(userID); err != nil {
		w.Header().Set(contentType, jsonType)
		w.WriteHeader(http.StatusInternalServerError)

		c.log.Error("ошибка в БД при регистрации пользователя", "err", err.Error())

		return
	}

	w.WriteHeader(http.StatusOK)
}

func (c *ChatHandler) DeleteHandler(w http.ResponseWriter, userID int64, userExist bool) {
	if !userExist {
		c.apiErrToResponse(w, dto.APIErrUserNotRegistered, http.StatusNotFound)

		return
	}

	err := c.transactor.WithTransaction(context.Background(), func(ctx context.Context) error {
		if err := c.userRepo.DeleteUser(ctx, userID); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		c.log.Error("ошибка в БД при удалении пользователя", "err", err.Error())

		return
	}

	w.WriteHeader(http.StatusOK)
}

func (c *ChatHandler) apiErrToResponse(w http.ResponseWriter, errAPI *dto.APIErrResponse, statusCode int) {
	w.Header().Set(contentType, jsonType)
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(errAPI)

	if err != nil {
		c.log.Error("ошибка при формировании JSON APIErrResponse", "err", err.Error())
	}
}
