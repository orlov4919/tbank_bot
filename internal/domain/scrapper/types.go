package scrapper

type User = int
type Link = string
type LinkState = string

type StackUpdate struct {
	Updates []LastUpdate `json:"items"`
}

type LastUpdate struct {
	UpdateTime int `json:"last_activity_date"`
}
