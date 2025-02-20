package updatesServer

import (
	"encoding/json"
	"fmt"
	"io"
	"linkTraccer/internal/domain/tgbot"
	"log"
	"net/http"
	"runtime"
)

const (
	contentType            = "Content-Type"
	jsonType               = "application/json"
	jsonError              = "json, который пришел в запросе не соответствует описанию LinkUpdate"
	methodError            = "нет доступа к запрашиваемому методу"
	methodErrorDescription = `Вы пытаетесь вызвать метод %s, который не поддерживает данный endpoint.
                              Поддерживаемые методы: POST`
)

// errors names
const (
	ReadBodyError = "ошибка при чтении тела запроса"
)

type UpdateServer struct {
	Server *http.Server
	TgBot  tgbot.BotService
}

func New(serverURL string, tgBot tgbot.BotService) *UpdateServer {
	return &UpdateServer{
		Server: &http.Server{
			Addr: serverURL,
		},
		TgBot: tgBot,
	}
}

func (s *UpdateServer) StartUpdatesService() error { // попробовать добавить handler и addr как два аргумента или только handler
	mux := http.NewServeMux()

	mux.HandleFunc("/updates", s.HandleLinkUpdates)

	s.Server.Handler = mux

	return s.Server.ListenAndServe()
}

func (s *UpdateServer) HandleLinkUpdates(w http.ResponseWriter, r *http.Request) {
	linkUpdates, err := io.ReadAll(r.Body)

	if err != nil {
		log.Println("Здесь наебнулся 1 ")
		errorResponse := NewApiErrorResponse(ReadBodyError, err.Error(), getStackTrace())
		jsonResponse, _ := json.Marshal(errorResponse)

		WriteInResponse(w, http.StatusBadRequest, jsonResponse)

		return
	}

	defer r.Body.Close()

	linkUpdatesJSON := &LinkUpdate{}

	if err = json.Unmarshal(linkUpdates, linkUpdatesJSON); err != nil {
		log.Println(string(linkUpdates))
		log.Println("Здесь наебнулся 2")

		errorResponse := NewApiErrorResponse(jsonError, err.Error(), getStackTrace())

		WriteInResponse(w, http.StatusBadRequest, errorResponse)

	} else if r.Method != http.MethodPost {
		log.Println("Здесь наебнулся 3")

		errorResponse := NewApiErrorResponse(methodError, fmt.Sprintf(methodErrorDescription, r.Method), getStackTrace())

		WriteInResponse(w, http.StatusBadRequest, errorResponse)

	} else {
		log.Println("Все с кайфом")

		WriteInResponse(w, http.StatusOK, nil)

		for _, userID := range linkUpdatesJSON.TgChatIds {
			go s.TgBot.SendMessage(userID, linkUpdatesJSON.Description)
		}
	}
}

func WriteInResponse(w http.ResponseWriter, HTTPStatus int, data any) {
	w.WriteHeader(HTTPStatus)

	if HTTPStatus == http.StatusOK {
		return
	}

	w.Header().Set(contentType, jsonType)

	if data != nil {

		err := json.NewEncoder(w).Encode(data)

		if err != nil {
			log.Println(fmt.Errorf("при формировании json ответа произошла ошибка: %w", err))
		}
	}
}

func getStackTrace() []string {
	stack := make([]uintptr, 32)
	length := runtime.Callers(2, stack[:]) // Пропускаем 2 вызова (Callers + getStackTrace)
	frames := runtime.CallersFrames(stack[:length])

	var trace []string
	for {
		frame, more := frames.Next()
		trace = append(trace, fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function))
		if !more {
			break
		}
	}
	return trace
}
