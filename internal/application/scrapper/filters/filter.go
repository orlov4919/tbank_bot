package filters

import (
	"linkTraccer/internal/application/scrapper/scrapservice"
	"linkTraccer/internal/domain/scrapper"
)

type UpdatesFilter struct {
	userRepo scrapservice.UserRepo
}

func New(repo scrapservice.UserRepo) *UpdatesFilter {
	return &UpdatesFilter{userRepo: repo}
}

func (f *UpdatesFilter) FilterByCreator(info *scrapper.LinkInfo, update *scrapper.LinkUpdate) ([]scrapper.User, error) {

	// одним sql запросом доставать тех кто не подходит по ссылке

}
