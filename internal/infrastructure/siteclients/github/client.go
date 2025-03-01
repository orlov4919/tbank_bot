package github

import (
	"encoding/json"
	"fmt"
	"linkTraccer/internal/domain/scrapper"
	"linkTraccer/internal/infrastructure/siteclients"
	"net/http"
	"net/url"
	"path"
	"strings"
)

const (
	pathLen        = 3
	clientName     = "GitHub"
	gitHubHost     = "github.com"
	repoCreaterInd = 1
	repoNameInd    = 2
	emptyArg       = ""
	issues         = "issues"
)

type Link = scrapper.Link
type LinkState = scrapper.LinkState

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
		basePath: "repos",
		client:   client,
	}
}

func (git *GitClient) CanTrack(link Link) bool {
	parsedLink, err := url.Parse(link)

	if err != nil {
		return false
	}

	pathArgs := strings.Split(parsedLink.Path, "/")

	if !git.StaticLinkCheck(parsedLink, pathArgs) {
		return false
	}

	q := url.Values{}

	q.Add("per_page", "1")
	q.Add("sort", "updated")

	reqURL := git.makeRequestURL(pathArgs[repoCreaterInd], pathArgs[repoNameInd], q)
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

	return resp.StatusCode != http.StatusOK
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

func (git *GitClient) LinkState(link Link) (LinkState, error) {
	parsedLink, err := url.Parse(link)

	if err != nil {
		return "", fmt.Errorf("в клиете %s при парсинге ссылки произошла ошибка: %w", clientName, err)
	}

	pathArgs := strings.Split(parsedLink.Path, "/")

	if !git.StaticLinkCheck(parsedLink, pathArgs) {
		return "", siteclients.NewErrClientCantTrackLink(link, clientName)
	}

	q := url.Values{}

	q.Add("per_page", "1")
	q.Add("sort", "updated")

	reqURL := git.makeRequestURL(pathArgs[repoCreaterInd], pathArgs[repoNameInd], q)
	req, err := http.NewRequest(http.MethodGet, reqURL.String(), http.NoBody)

	if err != nil {
		return "", fmt.Errorf("в клиете %s при формировании запроса произошла ошибка: %w", clientName, err)
	}

	req.Header.Add("Authorization", git.token)

	resp, err := git.client.Do(req)

	if err != nil {
		return "", fmt.Errorf("не смогли получить состояние ссылки, запрос кончился ошибкой :%w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", siteclients.NewErrBadRequestStatus("не смогли получить состояние ссылки", resp.StatusCode)
	}

	gitUpdates := make([]scrapper.LastGitUpdate, 0, 1)

	if err := json.NewDecoder(resp.Body).Decode(&gitUpdates); err != nil {
		return "", fmt.Errorf("в клиете %s при парсиге ответа произошла ошибка: %w", clientName, err)
	}

	if len(gitUpdates) == 0 {
		return "", nil
	}

	return gitUpdates[0].UpdateTime, nil
}

func (git *GitClient) makeRequestURL(repoAuthor, repo string, q url.Values) *url.URL {
	return &url.URL{
		Scheme:   git.scheme,
		Host:     git.host,
		Path:     path.Join(git.basePath, repoAuthor, repo, issues),
		RawQuery: q.Encode(),
	}
}
