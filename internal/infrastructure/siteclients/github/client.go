package github

import (
	"encoding/json"
	"fmt"
	"linkTraccer/internal/domain/scrapper"
	"linkTraccer/internal/infrastructure/siteclients"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	pathLen        = 3
	clientName     = "GitHub"
	gitHubHost     = "github.com"
	repoCreaterInd = 1
	repoNameInd    = 2
	emptyArg       = ""
	issue          = "Issue"
	pullRequest    = "Pull Request"
	maxTitleLen    = 200
)

type Link = scrapper.Link
type LinkUpdates = scrapper.LinkUpdates
type LinkUpdate = scrapper.LinkUpdate
type GitUpdates = scrapper.GitUpdates

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type GitClient struct {
	scheme   string
	basePath string
	host     string
	token    string
	client   HTTPClient
}

// при инициализации вводить api.github.com

func NewClient(host, token string, client HTTPClient) *GitClient {
	return &GitClient{
		scheme:   "https",
		host:     host,
		token:    token,
		basePath: "search/issues",
		client:   client,
	}
}

// УЖЕСТОЧИТЬ ПРОВЕРКИ НЕ ДОЛЖНАЯ ССЫЛКА https://github.com/orlov4919/test/issues прохоидить проверку, добавление новых тестов

func (git *GitClient) CanTrack(link Link) bool {
	parsedLink, err := url.Parse(link)

	if err != nil {
		return false
	}

	pathArgs := strings.Split(parsedLink.Path, "/")

	if !git.StaticLinkCheck(parsedLink, pathArgs) {
		return false
	}

	q := git.MakeQueryParams(pathArgs, time.Now().Truncate(time.Second))
	reqURL := git.makeRequestURL(q)

	req, err := http.NewRequest(http.MethodGet, reqURL.String(), http.NoBody)

	if err != nil {
		return false
	}

	req.Header.Add("Authorization", git.token)

	resp, err := git.client.Do(req)

	if err != nil {
		return false
	}

	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func (git *GitClient) StaticLinkCheck(parsedLink *url.URL, pathArgs []string) bool {
	cleanedHost := strings.TrimPrefix(parsedLink.Host, "www.")

	if cleanedHost != gitHubHost || parsedLink.Scheme != git.scheme {
		return false
	}

	if len(pathArgs) != pathLen {
		return false
	}

	if pathArgs[repoCreaterInd] == emptyArg || pathArgs[repoNameInd] == emptyArg {
		return false
	}

	return true
}

func (git *GitClient) LinkUpdates(link Link, updatesSince time.Time) (LinkUpdates, error) {
	parsedLink, err := url.Parse(link)

	updatesSince = updatesSince.Add(-time.Hour * 3)

	if err != nil {
		return nil, fmt.Errorf("в клиете %s при парсинге ссылки произошла ошибка: %w", clientName, err)
	}

	pathArgs := strings.Split(parsedLink.Path, "/")

	if !git.StaticLinkCheck(parsedLink, pathArgs) {
		return nil, siteclients.NewErrClientCantTrackLink(link, clientName)
	}

	q := git.MakeQueryParams(pathArgs, updatesSince)
	reqURL := git.makeRequestURL(q)
	req, err := http.NewRequest(http.MethodGet, reqURL.String(), http.NoBody)

	if err != nil {
		return nil, fmt.Errorf("в клиете %s при формировании запроса произошла ошибка: %w", clientName, err)
	}

	req.Header.Add("Authorization", git.token)

	resp, err := git.client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("не смогли получить состояние ссылки, запрос кончился ошибкой :%w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, siteclients.NewErrBadRequestStatus("не смогли получить состояние ссылки", resp.StatusCode)
	}

	gitUpdates := &GitUpdates{}

	if err := json.NewDecoder(resp.Body).Decode(&gitUpdates); err != nil {
		return nil, fmt.Errorf("в клиете %s при парсиге ответа произошла ошибка: %w", clientName, err)
	}

	return git.GitUpdatesToLinkUpdates(gitUpdates), nil
}

func (git *GitClient) GitUpdatesToLinkUpdates(gitUpdates *GitUpdates) LinkUpdates {
	var updateType string

	linkUpdates := make([]*LinkUpdate, 0, gitUpdates.Count)

	for _, update := range gitUpdates.Updates {

		if update.PullRequest.URL == "" {
			updateType = issue
		} else {
			updateType = pullRequest
		}

		createdTime, _ := time.Parse("2006-01-02T15:04:05Z", update.CreatedTime)
		createdTimeToMsk := createdTime.Add(time.Hour * 3)

		linkUpdates = append(linkUpdates, &LinkUpdate{
			Header:     updateType,
			UserName:   update.GitUser.Login,
			CreateTime: createdTimeToMsk.String(),
			Preview:    update.Title[:min(len(update.Title), maxTitleLen)],
		})
	}

	return linkUpdates
}

func (git *GitClient) MakeQueryParams(pathArgs []string, updatesSince time.Time) url.Values {
	q := url.Values{}

	q.Add("q", makeQueryString(pathArgs[repoCreaterInd], pathArgs[repoNameInd], updatesSince))
	q.Add("sort", "created")
	q.Add("per_page", "100")

	return q
}

func makeQueryString(repoAuthor, repo string, sinceTime time.Time) string {
	return "repo:" + repoAuthor + "/" + repo + " " + "created:>" + sinceTime.Format("2006-01-02T15:04:05Z")
}

func (git *GitClient) makeRequestURL(q url.Values) *url.URL {
	return &url.URL{
		Scheme:   git.scheme,
		Host:     git.host,
		Path:     git.basePath,
		RawQuery: q.Encode(),
	}
}

//https://api.github.com/search/issues?q=repo:orlov4919/test+created:%3E2011-01-01
