package scrapper

import "time"

type User = int64
type Link = string
type LinkState = string
type LinkID = int64
type Tag = string

type StackAnswers struct {
	Answers []StackAnswer `json:"items"`
}

type StackAnswer struct {
	UpdateTime int64  `json:"last_activity_date"`
	Title      string `json:"title"`
	Body       string `json:"body"`
	Owner      Owner  `json:"owner"`
}

type StackComments struct {
	Comments []StackComment `json:"items"`
}

type StackComment struct {
	UpdateTime int64  `json:"creation_date"`
	Title      string `json:"title"`
	Body       string `json:"body"`
	Owner      Owner  `json:"owner"`
}

type Owner struct {
	UserName string `json:"display_name"`
}

type LastGitUpdate struct {
	UpdateTime string `json:"updated_at"`
}

type LinkResponse struct {
	ID      int      `json:"id"`
	URL     Link     `json:"url"`
	Tags    []string `json:"tags"`
	Filters []string `json:"filters"`
}

type ListLinksResponse struct {
	Links []LinkResponse `json:"links"`
	Size  int            `json:"size"`
}

type AddLinkRequest struct {
	Link    string   `json:"link"`
	Tags    []string `json:"tags"`
	Filters []string `json:"filters"`
}

type RemoveLinkRequest struct {
	Link string `json:"link"`
}

type GitUpdates struct {
	Count   int         `json:"total_count"`
	Updates []GitUpdate `json:"items"`
}

type GitUpdate struct {
	GitUser     GitUser   `json:"user"`
	Title       string    `json:"title"`
	CreatedTime time.Time `json:"created_at"`
	PullRequest PR        `json:"pull_request"`
}

type GitUser struct {
	Login string `json:"login"`
}

type PR struct {
	URL string `json:"url"`
}

type LinkInfo struct {
	ID         LinkID
	URL        Link
	LastUpdate time.Time
}
