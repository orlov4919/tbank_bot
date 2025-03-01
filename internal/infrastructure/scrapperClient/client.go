package scrapperClient

import (
	"bytes"
	"encoding/json"
	"errors"
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

func New(client HTTPClient, host string) *ScrapperClient {
	return &ScrapperClient{
		host:           host,
		baseLinkPath:   "/links",
		baseTgChatPath: "/tg-chat",
		client:         client,
	}
}

func (s *ScrapperClient) RegUser(id ID) error {
	url := &url.URL{
		Scheme: "http",
		Host:   s.host,
		Path:   path.Join(s.baseTgChatPath, strconv.Itoa(id)),
	}

	req := &http.Request{
		Method: http.MethodPost,
		URL:    url,
	}

	resp, err := s.client.Do(req)

	if err != nil {
		return fmt.Errorf("Ошибка при вызове UserLinks: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("уже регнут")
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
			"Tg-Chat-Id": {strconv.Itoa(id)},
		},
	}

	resp, err := s.client.Do(req)

	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ошибка при вызове UserLinks: %w", err)
	}

	defer resp.Body.Close()

	listLinks := &scrapper.ListLinksResponse{}

	if err = json.NewDecoder(resp.Body).Decode(listLinks); err != nil {
		return nil, err
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
		log.Println("Ошибка при маршалинге")
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
			"Tg-Chat-Id":   {strconv.Itoa(id)},
			"Content-Type": {"application/json"},
		},
		Body: io.NopCloser(bytes.NewBuffer(removeLink)),
	}

	resp, err := s.client.Do(req)

	if err != nil {
		log.Println("Ошибка при запрос на сервер")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("Не смогли удалить ссылку")
	}

	return nil
}

func (s *ScrapperClient) AddLink(id ID, userCtx *tgbot.ContextData) error {

	log.Println("Добавляем ссылку")

	addLink, err := json.Marshal(&scrapper.AddLinkRequest{
		Link:    userCtx.URL,
		Tags:    userCtx.Tags,
		Filters: userCtx.Filters,
	})

	if err != nil {
		log.Println("Ошибка при маршалинге")
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
			"Tg-Chat-Id":   {strconv.Itoa(id)},
			"Content-Type": {"application/json"},
		},
		Body: io.NopCloser(bytes.NewBuffer(addLink)),
	}

	fmt.Println(url.String())

	resp, err := s.client.Do(req)

	if err != nil {
		log.Println("Ошибка при запрос на сервер")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("Не смогли добавить ссылку")
	}

	return nil
}
