package scrapperhandlers

import (
	"encoding/json"
	"linkTraccer/internal/domain/dto"
	"log"
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
}

func NewChatHandler(repo UserRepo) *ChatHandler {
	return &ChatHandler{
		userRepo: repo,
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

		err := json.NewEncoder(w).Encode(dto.NewApiErrResponse("неправильный формат данных в теле", err.Error(), []string{}))

		if err != nil {
			log.Println("Ошибка при формировании json ответа")
		}

		return
	}

	if userID < 0 {
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(dto.NewApiErrResponse("id error", "id < 0", []string{}))

		if err != nil {
			log.Println("Ошибка при формировании json ответа")
		}

		return
	}

	switch r.Method {
	case http.MethodPost:
		if c.userRepo.UserExist(userID) {
			w.WriteHeader(http.StatusBadRequest)

			err := json.NewEncoder(w).Encode(dto.NewApiErrResponse("id error", "id exist", []string{}))

			if err != nil {
				log.Println("Ошибка при формировании json ответа")
			}

			return
		}

		w.WriteHeader(http.StatusOK)

		err := c.userRepo.RegUser(userID)

		if err != nil {
			log.Println("Хранилище не согласовано")
		}

	case http.MethodDelete:
		if err := c.userRepo.DeleteUser(userID); err != nil {
			w.WriteHeader(http.StatusNotFound)

			err := json.NewEncoder(w).Encode(dto.NewApiErrResponse("id error", err.Error(), []string{}))

			if err != nil {
				log.Println("Ошибка при формировании json ответа")
			}
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}
}
