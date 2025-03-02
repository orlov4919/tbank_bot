package scrapperhandlers

import (
	"encoding/json"
	"linkTraccer/internal/domain/dto"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

const (
	contentType = "Content-Type"
	jsonType    = "application/json"
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
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set(contentType, jsonType)

	data := strings.TrimPrefix(r.URL.String(), "/tg-chat/")
	userID, err := strconv.Atoi(data)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		err := json.NewEncoder(w).Encode(dto.NewAPIErrResponse("неправильный формат данных в теле", err.Error(), []string{}))

		if err != nil {
			c.log.Debug("ошибка при энкодинге тела", "err", err.Error())
		}

		return
	}

	if userID < 0 {
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(dto.NewAPIErrResponse("id error", "id < 0", []string{}))

		if err != nil {
			c.log.Debug("ошибка при формировании json ответа", "err", err.Error())
		}

		return
	}

	switch r.Method {
	case http.MethodPost:
		if c.userRepo.UserExist(userID) {
			w.WriteHeader(http.StatusBadRequest)

			err := json.NewEncoder(w).Encode(dto.NewAPIErrResponse("id error", "id exist", []string{}))

			if err != nil {
				c.log.Debug("ошибка при формировании json ответа", "err", err.Error())
			}

			return
		}

		w.WriteHeader(http.StatusOK)

		if err := c.userRepo.RegUser(userID); err != nil {
			c.log.Info("ошибка при регистрации пользователя", "err", err.Error())
		}

	case http.MethodDelete:
		if err := c.userRepo.DeleteUser(userID); err != nil {
			w.WriteHeader(http.StatusNotFound)

			err := json.NewEncoder(w).Encode(dto.NewAPIErrResponse("id error", err.Error(), []string{}))

			if err != nil {
				c.log.Debug("ошибка при формировании json ответа", "err", err.Error())
			}
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}
}
