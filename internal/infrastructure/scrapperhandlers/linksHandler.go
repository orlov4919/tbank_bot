package scrapperhandlers

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	//"linkTraccer/internal/application/scrapper"
	"linkTraccer/internal/application/scrapper/scrapservice"
	"linkTraccer/internal/domain/dto"
	"linkTraccer/internal/domain/scrapper"
	"log/slog"
	"net/http"
	"strconv"
)

type LinkResponse = scrapper.LinkResponse
type ListLinksResponse = scrapper.ListLinksResponse
type UserRepo = scrapservice.UserRepo
type AddLinkRequest = scrapper.AddLinkRequest
type SiteClient = scrapservice.SiteClient
type RemoveLink = scrapper.RemoveLinkRequest

type LinkHandler struct {
	userRepo    UserRepo
	siteClients []SiteClient
	log         *slog.Logger
}

func NewLinkHandler(repo UserRepo, log *slog.Logger, clients ...SiteClient) *LinkHandler {
	return &LinkHandler{
		userRepo:    repo,
		siteClients: clients,
		log:         log,
	}
}

func (l *LinkHandler) HandleLinksChanges(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(contentType, jsonType)

	reqData, err := io.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		err := json.NewEncoder(w).Encode(dto.NewAPIErrResponse("ошибка тела запроса", err.Error(), []string{}))

		if err != nil {
			l.log.Info("Ошибка при формировании json ответа", "err", err)
		}

		return
	}

	user := r.Header.Get("Tg-Chat-Id")
	userID, err := strconv.ParseInt(user, 10, 64)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		err := json.NewEncoder(w).Encode(dto.NewAPIErrResponse("id error", "no id in req", []string{}))

		if err != nil {
			l.log.Info("Ошибка при формировании json ответа", "err", err)
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
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (l *LinkHandler) GetMethodHandler(w http.ResponseWriter, userID int64) {
	listLinksResponse := &ListLinksResponse{}
	userLinks, err := l.userRepo.AllUserLinks(userID)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		err := json.NewEncoder(w).Encode(dto.NewAPIErrResponse("id error", "id exist", []string{}))

		if err != nil {
			l.log.Info("Ошибка при формировании json ответа", "err", err)
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
		l.log.Info("Ошибка при формировании json ответа", "err", err)
	}
}

func (l *LinkHandler) PostMethodHandler(w http.ResponseWriter, userID int64, reqData []byte) {
	addLinkRequest := &AddLinkRequest{}

	fmt.Println()

	err := json.Unmarshal(reqData, addLinkRequest)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		err := json.NewEncoder(w).Encode(dto.NewAPIErrResponse("json err", err.Error(), []string{}))

		if err != nil {
			l.log.Info("Ошибка при формировании json ответа", "err", err)
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

	if !flag {

		w.WriteHeader(http.StatusBadRequest)

		err := json.NewEncoder(w).Encode(dto.NewAPIErrResponse("link err", "link not support", []string{}))

		if err != nil {
			l.log.Info("Ошибка при формировании json ответа", "err", err)
		}

		return
	}

	userTrackLink, err := l.userRepo.UserTrackLink(userID, addLinkRequest.Link)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		l.log.Info("ошибка при проверке отслеживания ссылки", "err", err)

		return
	}

	if userTrackLink {
		w.WriteHeader(http.StatusBadRequest)

		err := json.NewEncoder(w).Encode(dto.NewAPIErrResponse("already track", "user track this link", []string{}))

		if err != nil {
			l.log.Info("Ошибка при формировании json ответа", "err", err)
		}

		return
	}

	err = l.userRepo.TrackLink(userID, addLinkRequest.Link, time.Now().Truncate(time.Second))

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		l.log.Info("ошибка при добавлении ссылки", "err", err)

		return
	}

	w.WriteHeader(http.StatusOK)

	linkResponse := &LinkResponse{
		ID:      1,
		URL:     addLinkRequest.Link,
		Tags:    addLinkRequest.Tags,
		Filters: addLinkRequest.Filters,
	}

	if err := json.NewEncoder(w).Encode(linkResponse); err != nil {
		l.log.Debug("Ошибка при формировании json ответа", "err", err)
	}
}

func (l *LinkHandler) DeleteMethodHandler(w http.ResponseWriter, userID int64, reqData []byte) {
	removeLink := &RemoveLink{}
	err := json.Unmarshal(reqData, removeLink)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		err := json.NewEncoder(w).Encode(dto.NewAPIErrResponse("json error", "wrong json format", []string{}))

		if err != nil {
			l.log.Info("Ошибка при формировании json ответа", "err", err)
		}

		return
	}

	userTrackLink, err := l.userRepo.UserTrackLink(userID, removeLink.Link)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		l.log.Info("ошибка при проверке отслеживания ссылки", "err", err)

		return
	}

	if userTrackLink {

		err = l.userRepo.UntrackLink(userID, removeLink.Link)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			l.log.Info("ошибка при удалении ссылки", "err", err)

			return
		}

		w.WriteHeader(http.StatusOK)

		linkResponse := &LinkResponse{
			ID:      1,
			URL:     removeLink.Link,
			Tags:    []string{},
			Filters: []string{},
		}

		err := json.NewEncoder(w).Encode(linkResponse)

		if err != nil {
			l.log.Info("Ошибка при формировании json ответа", "err", err)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)

		err := json.NewEncoder(w).Encode(dto.NewAPIErrResponse("link err", "dont tracc link", []string{}))

		if err != nil {
			l.log.Info("Ошибка при формировании json ответа", "err", err)
		}
	}
}
