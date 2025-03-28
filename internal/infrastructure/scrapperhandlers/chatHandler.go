package scrapperhandlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"linkTraccer/internal/domain/dto"
	"log/slog"
	"net/http"
	"strconv"
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
	w.Header().Set(contentType, jsonType)

	userID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)

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

	userExist, err := c.userRepo.UserExist(userID)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		c.log.Info("ошибка при проверке пользователя в БД", "err", err.Error())

		return
	}

	switch r.Method {
	case http.MethodPost:

		if userExist {
			w.WriteHeader(http.StatusBadRequest)

			err := json.NewEncoder(w).Encode(dto.NewAPIErrResponse("id error", "id exist", []string{}))

			if err != nil {
				c.log.Debug("ошибка при формировании json ответа", "err", err.Error())
			}

			return
		}

		if err := c.userRepo.RegUser(userID); err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			c.log.Info("ошибка при регистрации пользователя", "err", err.Error())
		} else {
			w.WriteHeader(http.StatusOK)
		}

	case http.MethodDelete:

		if !userExist {
			w.WriteHeader(http.StatusNotFound)

			err := json.NewEncoder(w).Encode(dto.NewAPIErrResponse("id error", "нет чата", []string{}))

			if err != nil {
				c.log.Debug("ошибка при формировании json ответа", "err", err.Error())
			}

			return
		}

		if err := c.userRepo.DeleteUser(userID); err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			c.log.Info("ошибка при регистрации пользователя", "err", err.Error())
		} else {
			w.WriteHeader(http.StatusOK)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
