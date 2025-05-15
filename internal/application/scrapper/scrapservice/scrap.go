package scrapservice

import (
	"context"
	"fmt"
	"linkTraccer/internal/domain/scrapper"
	"log/slog"
	"sync"
	"time"
)

const (
	workersNum = 4
)

var (
	MoskowTime = time.FixedZone("UTC+3", 3*60*60)
)

type LinkPaginator interface {
	LinksBatch() ([]*scrapper.LinkInfo, error)
	HasLinks() bool
}

type UserRepo interface {
	NewLinksPaginator() LinkPaginator
	TrackLink(ctx context.Context, userID scrapper.User, link scrapper.Link, update time.Time) error
	ChangeLastCheckTime(link scrapper.Link, checkTime time.Time) error
	UsersWhoTrackLink(linkID scrapper.LinkID) ([]scrapper.User, error)
	AllUserLinks(userID scrapper.User) ([]scrapper.Link, error)
	UserTrackLink(userID scrapper.User, URL scrapper.Link) (bool, error)
	UntrackLink(user scrapper.User, link scrapper.Link) error
	UserExist(UserID scrapper.User) (bool, error)
	RegUser(UserID scrapper.User) error
	DeleteUser(ctx context.Context, user scrapper.User) error
}

type SiteClient interface {
	CanTrack(link scrapper.Link) bool
	LinkUpdates(link scrapper.Link, updatesSince time.Time) (scrapper.LinkUpdates, error)
}

type NotifyService interface {
	SendUpdates(linkInfo *scrapper.LinkInfo, linkUpdates scrapper.LinkUpdates) error
}

type FilterService interface {
	FilterByCreator()
}

type Transactor interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type Scrapper struct {
	userRepo      UserRepo
	siteClients   []SiteClient
	notifyService NotifyService
	log           *slog.Logger
}

func New(userRepo UserRepo, notifyService NotifyService, log *slog.Logger, siteClients ...SiteClient) *Scrapper {
	return &Scrapper{
		userRepo:      userRepo,
		notifyService: notifyService,
		siteClients:   siteClients,
		log:           log,
	}
}

func (scrap *Scrapper) LinksUpdates() {
	linksPaginator := scrap.userRepo.NewLinksPaginator()

	for linksPaginator.HasLinks() {
		links, err := linksPaginator.LinksBatch()
		if err != nil {
			scrap.log.Error("ошибка при получении батча ссылок", "err", err.Error())

			continue
		}

		linksChan := make(chan *scrapper.LinkInfo, len(links))

		go linksToChan(links, linksChan)

		wg := &sync.WaitGroup{}

		wg.Add(workersNum)

		for worker := 0; worker < workersNum; worker++ {
			go scrap.findUpdates(wg, linksChan)
		}

		wg.Wait()
	}
}

func linksToChan(links []*scrapper.LinkInfo, out chan<- *scrapper.LinkInfo) {
	for _, link := range links {
		out <- link
	}

	close(out)
}

func (scrap *Scrapper) findUpdates(wg *sync.WaitGroup, linksChan <-chan *scrapper.LinkInfo) {
	defer wg.Done()

	for linkInfo := range linksChan {
		for _, siteClient := range scrap.siteClients {

			if !siteClient.CanTrack(linkInfo.URL) {
				continue
			}

			t := time.Now().In(MoskowTime).Truncate(time.Second)

			linkUpdates, err := siteClient.LinkUpdates(linkInfo.URL, linkInfo.LastUpdate)
			if err != nil {
				scrap.log.Error("при получении обновлений ссылки произошла ошибка", "err", err.Error())
				break
			}

			if err = scrap.userRepo.ChangeLastCheckTime(linkInfo.URL, t); err != nil {
				scrap.log.Error("ошибка при изменении даты последней проверки ссылки", "err", err.Error())
				break
			}

			if len(linkUpdates) == 0 {
				break
			}

			scrap.log.Info(fmt.Sprintf("произошло %d обновлений по ссылке %s", len(linkUpdates), linkInfo.URL))

			 users, err := scrap.userRepo.UsersWhoTrackLink(linkInfo.ID) делаем это в фильтре

			for _, linkUpdate := range linkUpdates {
				users, err := scrap.filter.Users(linkInfo, linkUpdates)

				if err = scrap.notifyService.SendUpdates(linkInfo, linkUpdates); err != nil {
					scrap.log.Error("ошибка при отправке обновлений", "err", err.Error())
				}

			}
		}
	}
}
