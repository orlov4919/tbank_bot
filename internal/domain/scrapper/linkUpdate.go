package scrapper

type LinkUpdate struct {
	Header     string
	UserName   string
	CreateTime string
	Preview    string
}

type LinkUpdates = []*LinkUpdate
