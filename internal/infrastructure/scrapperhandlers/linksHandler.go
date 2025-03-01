package scrapperhandlers

import (
	"encoding/json"
	"io"
	"linkTraccer/internal/application/scrapperService"
	"linkTraccer/internal/domain/dto"
	"linkTraccer/internal/domain/scrapper"
	"log"
	"net/http"
	"strconv"
)

type LinkResponse = scrapper.LinkResponse
type ListLinksResponse = scrapper.ListLinksResponse
type UserRepo = scrapperService.UserRepo
type AddLinkRequest = scrapper.AddLinkRequest
type SiteClient = scrapperService.SiteClient
type RemoveLink = scrapper.RemoveLinkRequest

type LinkHandler struct {
	userRepo    UserRepo
	siteClients []SiteClient
}

func NewLinkHandler(repo UserRepo, clients ...SiteClient) *LinkHandler {
	return &LinkHandler{
		userRepo:    repo,
		siteClients: clients,
	}
}

func (l *LinkHandler) HandleLinksChanges(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete && r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set(contentType, jsonType)

	reqData, err := io.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		err := json.NewEncoder(w).Encode(dto.NewApiErrResponse("ошибка тела запроса", err.Error(), []string{}))

		if err != nil {
			log.Println("Ошибка при формировании json ответа")
		}

		return
	}

	user := r.Header.Get("Tg-Chat-Id")
	userID, err := strconv.Atoi(user)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		err := json.NewEncoder(w).Encode(dto.NewApiErrResponse("id error", "no id in req", []string{}))

		if err != nil {
			log.Println("Ошибка при формировании json ответа")
		}

		return
	}

	switch r.Method {
	case http.MethodGet:
		l.GetMethodHandler(w, userID)

	case http.MethodPost:
		l.PostMethodHandler(w, userID, reqData)

	case http.MethodDelete:
		l.DeleteMethodHandler(w, userID, reqData)
	}
}

func (l *LinkHandler) GetMethodHandler(w http.ResponseWriter, userID int) {
	listLinksResponse := &ListLinksResponse{}
	userLinks, err := l.userRepo.AllUserLinks(userID)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		err := json.NewEncoder(w).Encode(dto.NewApiErrResponse("id error", "id exist", []string{}))

		if err != nil {
			log.Println("Ошибка при формировании json ответа")
		}

		return
	}

	w.WriteHeader(http.StatusOK)

	for ind, link := range userLinks {
		linkResponse := LinkResponse{
			ID:      ind,
			URL:     link,
			Tags:    []string{},
			Filters: []string{}}

		listLinksResponse.Links = append(listLinksResponse.Links, linkResponse)
	}

	listLinksResponse.Size = len(listLinksResponse.Links)
	err = json.NewEncoder(w).Encode(listLinksResponse)

	if err != nil {
		log.Println("Ошибка при формировании json ответа")
	}
}

func (l *LinkHandler) PostMethodHandler(w http.ResponseWriter, userID int, reqData []byte) {
	addLinkRequest := &AddLinkRequest{}
	err := json.Unmarshal(reqData, addLinkRequest)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		err := json.NewEncoder(w).Encode(dto.NewApiErrResponse("json err", "wrong json format", []string{}))

		if err != nil {
			log.Println("Ошибка при формировании json ответа")
		}

		return
	}

	if l.userRepo.UserTrackLink(userID, addLinkRequest.Link) {
		w.WriteHeader(http.StatusBadRequest)

		err := json.NewEncoder(w).Encode(dto.NewApiErrResponse("already track", "user track this link", []string{}))

		if err != nil {
			log.Println("Ошибка при формировании json ответа")
		}

		return
	}

	flag := false

	for _, client := range l.siteClients {
		if client.CanTrack(addLinkRequest.Link) {
			flag = true
			break
		}
	}

	if flag {
		w.WriteHeader(http.StatusOK)

		err = l.userRepo.TrackLink(userID, addLinkRequest.Link, "") // Добавить состояние константой

		if err != nil {
			log.Println("При добавлении ссылки произошла ошибка")
		}

		linkResponse := &LinkResponse{
			ID:      1,
			URL:     addLinkRequest.Link,
			Tags:    addLinkRequest.Tags,
			Filters: addLinkRequest.Filters,
		}

		err := json.NewEncoder(w).Encode(linkResponse)

		if err != nil {
			log.Println("Ошибка при формировании json ответа")
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)

		err := json.NewEncoder(w).Encode(dto.NewApiErrResponse("link err", "link not support", []string{}))

		if err != nil {
			log.Println("Ошибка при формировании json ответа")
		}
	}
}

func (l *LinkHandler) DeleteMethodHandler(w http.ResponseWriter, userID int, reqData []byte) {
	removeLink := &RemoveLink{}
	err := json.Unmarshal(reqData, removeLink)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		err := json.NewEncoder(w).Encode(dto.NewApiErrResponse("json error", "wrong json format", []string{}))

		if err != nil {
			log.Println("Ошибка при формировании json ответа")
		}

		return
	}

	if l.userRepo.UserTrackLink(userID, removeLink.Link) {
		w.WriteHeader(http.StatusOK)

		err = l.userRepo.UntrackLink(userID, removeLink.Link)

		if err != nil {
			log.Println("При удалении ссылки произошла ошибка")
		}

		linkResponse := &LinkResponse{
			ID:      1,
			URL:     removeLink.Link,
			Tags:    []string{},
			Filters: []string{},
		}

		err := json.NewEncoder(w).Encode(linkResponse)

		if err != nil {
			log.Println("Ошибка при формировании json ответа")
		}
	} else {
		w.WriteHeader(http.StatusNotFound)

		err := json.NewEncoder(w).Encode(dto.NewApiErrResponse("link err", "dont tracc link", []string{}))

		if err != nil {
			log.Println("Ошибка при формировании json ответа")
		}
	}
}
