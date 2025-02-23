package stackoverflow

import (
	"encoding/json"
	"errors"
	"linkTraccer/internal/domain/scrapper"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
)

const (
	APIVersion = "/2.3"
)

const (
	maxPathLen        = 5
	indEmptyElement   = 0
	indQuestions      = 1
	indQuestionId     = 2
	stackoverflowHost = "stackoverflow.com"
)

type Link = scrapper.Link
type LinkState = scrapper.LinkState

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type StackClient struct {
	scheme   string
	basePath string
	host     string
	client   HTTPClient
}

// при инициализации вводить api.stackexchange.com

func NewClient(host string, client HTTPClient) *StackClient {

	return &StackClient{
		scheme:   "https",
		host:     host,
		basePath: path.Join(APIVersion, "questions"),
		client:   client,
	}
}

func (stack *StackClient) CanTrack(link Link) bool {
	parsedLink, err := url.Parse(link)

	if err != nil {
		return false
	}

	pathArgs := strings.Split(parsedLink.Path, "/")

	if !stack.StaticLinkCheck(parsedLink, pathArgs) {
		return false
	}

	q := url.Values{}
	q.Add("site", "stackoverflow")

	reqURL := stack.makeRequestURL(pathArgs[indQuestionId], q)
	req, err := http.NewRequest(http.MethodGet, reqURL.String(), nil)

	if err != nil {
		return false
	}

	resp, err := stack.client.Do(req)
	defer resp.Body.Close()

	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}

	return true
}

func (stack *StackClient) StaticLinkCheck(parsedLink *url.URL, pathArgs []string) bool {

	cleanedHost := strings.TrimPrefix(parsedLink.Host, "www.")

	if cleanedHost != stackoverflowHost || parsedLink.Scheme != stack.scheme {
		return false
	}

	if len(pathArgs) > maxPathLen || len(pathArgs) < 3 {
		return false
	}

	questionID, err := strconv.Atoi(pathArgs[2])

	if pathArgs[indEmptyElement] != "" || pathArgs[indQuestions] != "questions" || err != nil || questionID < 1 {
		return false
	}

	return true
}

func (stack *StackClient) LinkState(link Link) (LinkState, error) {
	parsedLink, err := url.Parse(link)

	if err != nil {
		return "", errors.New("Не правильно передана ссылка") // убрать
	}

	pathArgs := strings.Split(parsedLink.Path, "/")

	if !stack.StaticLinkCheck(parsedLink, pathArgs) {
		return "", errors.New("Не правильно передана ссылка") // убрать
	}

	q := url.Values{}
	q.Add("site", "stackoverflow")

	reqURL := stack.makeRequestURL(pathArgs[indQuestionId], q)
	req, err := http.NewRequest(http.MethodGet, reqURL.String(), nil)

	if err != nil {
		return "", errors.New("Ошибка при создании запроса")
	}

	resp, err := stack.client.Do(req)
	defer resp.Body.Close()

	if err != nil || resp.StatusCode != http.StatusOK {
		return "", errors.New("Запрос закончился ошибкой")
	}

	stackUpdate := &scrapper.StackUpdate{}

	if err := json.NewDecoder(resp.Body).Decode(stackUpdate); err != nil || len(stackUpdate.Updates) == 0 {
		return "", errors.New("Запрос закончился ошибкой")
	}

	return strconv.Itoa(stackUpdate.Updates[0].UpdateTime), nil
}

func (stack *StackClient) makeRequestURL(questionID string, q url.Values) *url.URL {
	return &url.URL{
		Scheme:   stack.scheme,
		Host:     stack.host,
		Path:     path.Join(stack.basePath, questionID),
		RawQuery: q.Encode(),
	}
}
