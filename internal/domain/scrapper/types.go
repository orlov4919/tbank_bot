package scrapper

type User = int
type Link = string
type LinkState = string

type StackUpdate struct {
	Updates []LastStackUpdate `json:"items"`
}

type LastStackUpdate struct {
	UpdateTime int `json:"last_activity_date"`
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
