package stackoverflow

import (
	"encoding/json"
	"fmt"
	"html"
	"linkTraccer/internal/domain/scrapper"
	"linkTraccer/internal/infrastructure/siteclients"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
)

const (
	APIVersion     = "/2.3"
	clientName     = "stackoverflow"
	answers        = "answers"
	comments       = "comments"
	answersFilter  = "!WWsh2-5LBtfz3hYj8MwV0S(v9oKR1U5(xsaX_2a"
	commentsFilter = "!szx-Dsx)YFm7RenuUsIW(gxHfTtAMj8"
	titleFiler     = "!)riR7ZAnK8mK6ZjITNAx"
	stackoverflow  = "stackoverflow"
	site           = "site"
	fromDate       = "fromdate"
	filter         = "filter"
)

const (
	minPathLen        = 3
	maxPathLen        = 4
	maxPreviewLen     = 200
	indEmptyElement   = 0
	indQuestions      = 1
	indQuestionID     = 2
	stackoverflowHost = "stackoverflow.com"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type StackClient struct {
	scheme     string
	basePath   string
	host       string
	client     HTTPClient
	strCleaner func(s string) string
}

// при инициализации вводить api.stackexchange.com

func NewClient(host string, client HTTPClient, strCleaner func(string) string) *StackClient {
	return &StackClient{
		scheme:     "https",
		host:       host,
		basePath:   path.Join(APIVersion, "questions"),
		client:     client,
		strCleaner: strCleaner,
	}
}

func (stack *StackClient) CanTrack(link scrapper.Link) bool {
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

	reqURL := stack.requestURL(pathArgs[indQuestionID], answers, q)
	req, err := http.NewRequest(http.MethodGet, reqURL.String(), http.NoBody)

	if err != nil {
		return false
	}

	resp, err := stack.client.Do(req)

	if err != nil {
		return false
	}

	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func (stack *StackClient) StaticLinkCheck(parsedLink *url.URL, pathArgs []string) bool {
	cleanedHost := strings.TrimPrefix(parsedLink.Host, "www.")

	if cleanedHost != stackoverflowHost || parsedLink.Scheme != stack.scheme {
		return false
	}

	if len(pathArgs) > maxPathLen || len(pathArgs) < minPathLen {
		return false
	}

	questionID, err := strconv.Atoi(pathArgs[indQuestionID])

	if err != nil || pathArgs[indEmptyElement] != "" || pathArgs[indQuestions] != "questions" || questionID < 1 {
		return false
	}

	return true
}

func (stack *StackClient) LinkUpdates(link scrapper.Link, since time.Time) (scrapper.LinkUpdates, error) {
	var questionTitle string

	since = since.Add(-time.Hour * 4)

	parsedLink, err := url.Parse(link)

	if err != nil {
		return nil, fmt.Errorf("в клиете %s при парсинге ссылки произошла ошибка: %w", clientName, err)
	}

	pathArgs := strings.Split(parsedLink.Path, "/")

	if !stack.StaticLinkCheck(parsedLink, pathArgs) {
		return nil, siteclients.NewErrClientCantTrackLink(link, clientName)
	}

	newAnswers, err := stack.NewAnswers(pathArgs[indQuestionID], since)

	if err != nil {
		return nil, fmt.Errorf("ошибка при получении новых ответов: %w", err)
	}

	newComments, err := stack.NewComments(pathArgs[indQuestionID], since)

	if err != nil {
		return nil, fmt.Errorf("ошибка при получении новых комментариев: %w", err)
	}

	if len(newAnswers.Items) == 0 && len(newComments.Items) != 0 {
		questionTitle, err = stack.getQuestionTitle(pathArgs)
	}

	if len(newAnswers.Items) != 0 {
		questionTitle = newAnswers.Items[0].Title
	}

	if err != nil {
		return nil, err
	}

	return stack.mergeUpdate(newAnswers, newComments, questionTitle), nil
}

func (stack *StackClient) mergeUpdate(answers *scrapper.StackAnswers, comments *scrapper.StackComments, title string) scrapper.LinkUpdates {
	linkUpdates := make(scrapper.LinkUpdates, 0, len(answers.Items)+len(comments.Items))

	for _, update := range answers.Items {
		linkUpdates = append(linkUpdates, &scrapper.LinkUpdate{
			Header:     title,
			UserName:   update.Owner.UserName,
			CreateTime: time.Unix(update.UpdateTime, 0).Format("15:04:05 02-01-2006"),
			Preview:    stack.strCleaner(update.Body),
		})
	}

	for _, update := range comments.Items {
		linkUpdates = append(linkUpdates, &scrapper.LinkUpdate{
			Header:     title,
			UserName:   update.Owner.UserName,
			CreateTime: time.Unix(update.UpdateTime, 0).Format("15:04:05 02-01-2006"),
			Preview:    stack.strCleaner(update.Body),
		})
	}

	return linkUpdates
}

func (stack *StackClient) NewUpdate(req *http.Request, update any) error {
	resp, err := stack.client.Do(req)

	if err != nil {
		return siteclients.NewErrNetwork(clientName, req.URL.String(), err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return siteclients.NewErrBadRequestStatus("не смогли получить состояние ссылки", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(update); err != nil {
		return fmt.Errorf("в клиете %s при парсиге ответа произошла ошибка: %w", clientName, err)
	}

	return nil
}

func (stack *StackClient) NewAnswers(questionID string, since time.Time) (*scrapper.StackAnswers, error) {
	q := url.Values{}

	q.Add(site, stackoverflow)
	q.Add(fromDate, strconv.FormatInt(since.Unix(), 10))
	q.Add(filter, answersFilter)

	reqURL := stack.requestURL(questionID, answers, q)

	req, err := http.NewRequest(http.MethodGet, reqURL.String(), http.NoBody)

	if err != nil {
		return nil, fmt.Errorf("в клиете %s при формировании запроса произошла ошибка: %w", clientName, err)
	}

	newAnswers := &scrapper.StackAnswers{}

	if err = stack.NewUpdate(req, newAnswers); err != nil {
		return nil, fmt.Errorf("ошибка при получении новых ответов: %w", err)
	}

	return newAnswers, nil
}

func (stack *StackClient) NewComments(questionID string, since time.Time) (*scrapper.StackComments, error) {
	q := url.Values{}

	q.Add(site, stackoverflow)
	q.Add(fromDate, strconv.FormatInt(since.Unix(), 10))
	q.Add(filter, commentsFilter)

	reqURL := stack.requestURL(questionID, comments, q)
	req, err := http.NewRequest(http.MethodGet, reqURL.String(), http.NoBody)

	if err != nil {
		return nil, fmt.Errorf("в клиете %s при формировании запроса произошла ошибка: %w", clientName, err)
	}

	newComments := &scrapper.StackComments{}

	if err = stack.NewUpdate(req, newComments); err != nil {
		return nil, fmt.Errorf("ошибка при получении новых комментариев: %w", err)
	}

	return newComments, nil
}

func (stack *StackClient) getQuestionTitle(pathArgs []string) (string, error) {
	q := url.Values{}

	q.Add(site, stackoverflow)
	q.Add(filter, titleFiler)

	reqURL := &url.URL{
		Scheme:   stack.scheme,
		Host:     stack.host,
		Path:     path.Join(stack.basePath, pathArgs[indQuestionID]),
		RawQuery: q.Encode(),
	}

	req, err := http.NewRequest(http.MethodGet, reqURL.String(), http.NoBody)

	if err != nil {
		return "", fmt.Errorf("в клиете %s при формировании запроса произошла ошибка: %w", clientName, err)
	}

	answers := &scrapper.StackAnswers{}

	if err = stack.NewUpdate(req, answers); err != nil {
		return "", fmt.Errorf("ошибка при получении заголовка вопроса: %w", err)
	}

	return answers.Items[0].Title, nil
}

func (stack *StackClient) requestURL(questionID, update string, q url.Values) *url.URL {
	return &url.URL{
		Scheme:   stack.scheme,
		Host:     stack.host,
		Path:     path.Join(stack.basePath, questionID, update),
		RawQuery: q.Encode(),
	}
}

func HTMLStrCleaner(maxPreviewLen int) func(s string) string {
	p := bluemonday.StripTagsPolicy()

	return func(s string) string {
		s = html.UnescapeString(p.Sanitize(s))

		if len(s) > maxPreviewLen {
			s = s[:maxPreviewLen]
		}

		return s
	}
}
