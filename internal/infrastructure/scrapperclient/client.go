package scrapperclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"linkTraccer/internal/domain/scrapper"
	"linkTraccer/internal/domain/tgbot"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
)

type Link = tgbot.Link
type ID = tgbot.ID

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type ScrapperClient struct {
	host           string
	baseLinkPath   string
	baseTgChatPath string
	client         HTTPClient
}

func New(client HTTPClient, host, port string) *ScrapperClient {
	return &ScrapperClient{
		host:           host + port,
		baseLinkPath:   "/links",
		baseTgChatPath: "/tg-chat",
		client:         client,
	}
}

func (s *ScrapperClient) RegUser(id ID) error {
	url := &url.URL{
		Scheme: "http",
		Host:   s.host,
		Path:   path.Join(s.baseTgChatPath, strconv.FormatInt(id, 10)),
	}

	req := &http.Request{
		Method: http.MethodPost,
		URL:    url,
	}

	resp, err := s.client.Do(req)

	if err != nil {
		return fmt.Errorf("запрос регистрации пользователя закончился ошибкой: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return NewErrBadRequestStatus("не получилось зарегистрировать юзера", resp.StatusCode)
	}

	return nil
}

func (s *ScrapperClient) UserLinks(id ID) ([]Link, error) {
	url := &url.URL{
		Scheme: "http",
		Host:   s.host,
		Path:   s.baseLinkPath,
	}

	req := &http.Request{
		Method: http.MethodGet,
		URL:    url,
		Header: map[string][]string{
			"Tg-Chat-Id": {strconv.FormatInt(id, 10)},
		},
	}

	resp, err := s.client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("во время выполнения запроса на получение ссылок возникла ошибка: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewErrBadRequestStatus("не смогли получить ссылки пользователя", resp.StatusCode)
	}

	listLinks := &scrapper.ListLinksResponse{}

	if err = json.NewDecoder(resp.Body).Decode(listLinks); err != nil {
		return nil, fmt.Errorf("не смогли десериализовать ссылки пользователя: %w ", err)
	}

	links := make([]Link, 0, listLinks.Size)

	for _, link := range listLinks.Links {
		links = append(links, link.URL)
	}

	return links, nil
}

func (s *ScrapperClient) RemoveLink(id ID, link Link) error {
	removeLink, err := json.Marshal(&scrapper.RemoveLinkRequest{Link: link})

	if err != nil {
		return fmt.Errorf("ошибка при маршалинге объекта, для удалениия ссылки: %w", err)
	}

	url := &url.URL{
		Scheme: "http",
		Host:   s.host,
		Path:   s.baseLinkPath,
	}

	req := &http.Request{
		Method: http.MethodDelete,
		URL:    url,
		Header: map[string][]string{
			"Tg-Chat-Id":   {strconv.FormatInt(id, 10)},
			"Content-Type": {"application/json"},
		},
		Body: io.NopCloser(bytes.NewBuffer(removeLink)),
	}

	resp, err := s.client.Do(req)

	if err != nil {
		return fmt.Errorf("запрос на удаление ссылки пользователя, закончился ошибкой :%w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return NewErrBadRequestStatus("не смогли удалить ссылку пользователя", resp.StatusCode)
	}

	return nil
}

func (s *ScrapperClient) AddLink(id ID, userCtx *tgbot.ContextData) error {
	addLink, err := json.Marshal(&scrapper.AddLinkRequest{
		Link:    userCtx.URL,
		Tags:    userCtx.Tags,
		Filters: userCtx.Filters,
	})

	if err != nil {
		return fmt.Errorf("не получилось добавить ссылку, ошибка при маршалинге: %w", err)
	}

	url := &url.URL{
		Scheme: "http",
		Host:   s.host,
		Path:   s.baseLinkPath,
	}

	req := &http.Request{
		Method: http.MethodPost,
		URL:    url,
		Header: map[string][]string{
			"Tg-Chat-Id":   {strconv.FormatInt(id, 10)},
			"Content-Type": {"application/json"},
		},
		Body: io.NopCloser(bytes.NewBuffer(addLink)),
	}

	resp, err := s.client.Do(req)

	if err != nil {
		log.Println("ошибка при выполнении запроса")
		return fmt.Errorf("запрос на добавление ссылки пользователя, закончился ошибкой :%w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Не верный статус")
		return NewErrBadRequestStatus("не смогли добавить ссылку пользователя", resp.StatusCode)
	}

	return nil
}
