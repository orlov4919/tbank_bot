package scraphandlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"linkTraccer/internal/application/scrapper/scrapservice"
	"linkTraccer/internal/domain/dto"
	"linkTraccer/internal/domain/scrapper"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

var (
	MoskowTime = time.FixedZone("UTC+3", 3*60*60)
)

type LinkResponse = scrapper.LinkResponse
type ListLinksResponse = scrapper.ListLinksResponse
type UserRepo = scrapservice.UserRepo
type AddLinkRequest = scrapper.AddLinkRequest
type SiteClient = scrapservice.SiteClient
type RemoveLink = scrapper.RemoveLinkRequest
type Transactor = scrapservice.Transactor

type LinkHandler struct {
	userRepo    UserRepo
	transactor  Transactor
	siteClients []SiteClient
	log         *slog.Logger
}

func NewLinkHandler(repo UserRepo, transactor Transactor, log *slog.Logger, clients ...SiteClient) *LinkHandler {
	return &LinkHandler{
		userRepo:    repo,
		transactor:  transactor,
		siteClients: clients,
		log:         log,
	}
}

func (l *LinkHandler) HandleLinksChanges(w http.ResponseWriter, r *http.Request) {
	reqData, err := io.ReadAll(r.Body)

	if err != nil {
		l.apiErrToResponse(w, dto.APIErrCantReadBody, http.StatusBadRequest)

		return
	}

	defer r.Body.Close()

	userID, err := strconv.ParseInt(r.Header.Get("Tg-Chat-Id"), 10, 64)

	if err != nil {
		l.apiErrToResponse(w, dto.APIErrIDNotNum, http.StatusBadRequest)

		return
	}

	if userID < 0 {
		l.apiErrToResponse(w, dto.APIErrNegativeID, http.StatusBadRequest)

		return
	}

	userExist, err := l.userRepo.UserExist(userID)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		if r.URL != nil {
			l.log.Error(
				fmt.Sprintf("обработка запроса %s закончилась ошибкой, при проверке пользователя в БД", r.URL.Path),
				"err", err.Error())
		}

		return
	}

	if !userExist {
		l.apiErrToResponse(w, dto.APIErrUserNotRegistered, http.StatusBadRequest)

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
		w.WriteHeader(http.StatusInternalServerError)

		if err != nil {
			l.log.Error(
				fmt.Sprintf("ошибка в БД при получении всех ссылок пользователя %d", userID),
				"err", err.Error())
		}

		return
	}

	w.WriteHeader(http.StatusOK)

	listLinksResponse.Size = len(userLinks)
	listLinksResponse.Links = make([]LinkResponse, 0, listLinksResponse.Size)

	for ind, link := range userLinks {
		linkResponse := LinkResponse{
			ID:      ind,
			URL:     link,
			Tags:    []string{},
			Filters: []string{}}

		listLinksResponse.Links = append(listLinksResponse.Links, linkResponse)
	}

	if err = json.NewEncoder(w).Encode(listLinksResponse); err != nil {
		l.log.Error(fmt.Sprintf("ошибка при формировании JSON всех ссылок пользователя %d", userID),
			"err", err)
	}
}

func (l *LinkHandler) PostMethodHandler(w http.ResponseWriter, userID int64, reqData []byte) {
	addLinkRequest := &AddLinkRequest{}

	if err := json.Unmarshal(reqData, addLinkRequest); err != nil {
		l.apiErrToResponse(w, dto.APIErrBadJSON, http.StatusBadRequest)
		return
	}

	canTrackLink := false

	for _, client := range l.siteClients {
		if client.CanTrack(addLinkRequest.Link) {
			canTrackLink = true
			break
		}
	}

	if !canTrackLink {
		l.apiErrToResponse(w, dto.APIErrBadLink, http.StatusBadRequest)
		return
	}

	userTrackLink, err := l.userRepo.UserTrackLink(userID, addLinkRequest.Link)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		l.log.Error(fmt.Sprintf("ошибка в БД при проверке, отслеживает пользователь %d ссылку %s",
			userID, addLinkRequest.Link), "err", err)

		return
	}

	if userTrackLink {
		l.apiErrToResponse(w, dto.APIErrDuplicateLink, http.StatusBadRequest)
		return
	}

	err = l.transactor.WithTransaction(context.Background(), func(ctx context.Context) error {
		err = l.userRepo.TrackLink(ctx, userID, addLinkRequest.Link, time.Now().In(MoskowTime).Truncate(time.Second))

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		l.log.Error(fmt.Sprintf("ошибка при добавлении в БД отслеживания пользователем %d ссылки %s",
			userID, addLinkRequest.Link),
			"err", err)

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
		l.log.Error("Ошибка при формировании JSON ответа, подтверждающего добавление новой ссылки ", "err", err)
	}
}

func (l *LinkHandler) DeleteMethodHandler(w http.ResponseWriter, userID int64, reqData []byte) {
	removeLink := &RemoveLink{}

	if err := json.Unmarshal(reqData, removeLink); err != nil {
		l.apiErrToResponse(w, dto.APIErrBadJSON, http.StatusBadRequest)

		return
	}

	userTrackLink, err := l.userRepo.UserTrackLink(userID, removeLink.Link)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		l.log.Error(fmt.Sprintf("ошибка в БД при проверке отслеживает пользователь %d ссылку %s",
			userID, removeLink.Link),
			"err", err)

		return
	}

	if !userTrackLink {
		l.apiErrToResponse(w, dto.APIErrNotTrackLink, http.StatusNotFound)

		return
	}

	err = l.userRepo.UntrackLink(userID, removeLink.Link)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		l.log.Error(fmt.Sprintf("ошибка в БД при удалении у пользователя %d ссылки %s", userID, removeLink.Link),
			"err", err)

		return
	}

	w.WriteHeader(http.StatusOK)

	linkResponse := &LinkResponse{
		ID:  1,
		URL: removeLink.Link,
	}

	if err = json.NewEncoder(w).Encode(linkResponse); err != nil {
		l.log.Error("Ошибка при формировании JSON ответа, подтверждающего удаление ссылки", "err", err)
	}
}

func (l *LinkHandler) apiErrToResponse(w http.ResponseWriter, errAPI *dto.APIErrResponse, statusCode int) {
	w.Header().Set(contentType, jsonType)
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(errAPI)

	if err != nil {
		l.log.Error("ошибка при формировании JSON APIErrResponse", "err", err.Error())
	}
}
