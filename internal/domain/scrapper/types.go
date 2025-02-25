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
