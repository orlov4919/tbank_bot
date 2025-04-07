package scrapservice

import (
	"context"
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

type User = scrapper.User
type Link = scrapper.Link
type LinkState = scrapper.LinkState
type LinkInfo = scrapper.LinkInfo
type LinkUpdates = scrapper.LinkUpdates
type LinkID = scrapper.LinkID
type LinkUpdate = scrapper.LinkUpdate

type LinkPaginator interface {
	LinksBatch() ([]*LinkInfo, error)
	HasLinks() bool
}

type UserRepo interface {
	NewLinksPaginator() LinkPaginator
	TrackLink(ctx context.Context, userID User, link Link, update time.Time) error
	ChangeLastCheckTime(link Link, checkTime time.Time) error
	UsersWhoTrackLink(linkID LinkID) ([]User, error)
	AllUserLinks(userID User) ([]Link, error)
	UserTrackLink(userID User, URL Link) (bool, error)
	UntrackLink(user User, link Link) error
	UserExist(UserID User) (bool, error)
	RegUser(UserID User) error
	DeleteUser(ctx context.Context, user User) error
}

type SiteClient interface {
	CanTrack(link Link) bool
	LinkUpdates(link Link, updatesSince time.Time) (LinkUpdates, error)
}

type NotifyService interface {
	SendUpdates(linkInfo *LinkInfo, linkUpdates LinkUpdates) error
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

func (scrap *Scrapper) CheckLinksUpdates() {
	linksPaginator := scrap.userRepo.NewLinksPaginator()

	for linksPaginator.HasLinks() {
		links, err := linksPaginator.LinksBatch()

		if err != nil {
			scrap.log.Error("ошибка при батчинге ссылок", "err", err.Error())

			continue
		}

		linksChan := make(chan *LinkInfo, len(links))

		go linksToChan(links, linksChan)

		wg := &sync.WaitGroup{}

		wg.Add(workersNum)

		for worker := 0; worker < workersNum; worker++ {
			scrap.checkLinksUpdates(wg, linksChan)
		}

		wg.Wait()
	}
}

func linksToChan(links []*LinkInfo, out chan<- *LinkInfo) {
	for _, link := range links {
		out <- link
	}

	close(out)
}

func (scrap *Scrapper) checkLinksUpdates(wg *sync.WaitGroup, linksChan <-chan *LinkInfo) {
	defer wg.Done()

	for linkInfo := range linksChan {
		for _, siteClient := range scrap.siteClients {

			if !siteClient.CanTrack(linkInfo.URL) {
				continue
			}

			t := time.Now().In(MoskowTime).Truncate(time.Second)

			linkUpdates, err := siteClient.LinkUpdates(linkInfo.URL, linkInfo.LastUpdate)

			if err != nil {
				scrap.log.Info("при получении состояния ссылки произошла ошибка", "err", err.Error())
				break
			}

			if err = scrap.userRepo.ChangeLastCheckTime(linkInfo.URL, t); err != nil {
				scrap.log.Info("ошибка при изменении даты последней проверки ссылки", "err", err.Error())
				break
			}

			if len(linkUpdates) == 0 {
				break
			}

			scrap.log.Info("ссылка " + linkInfo.URL + " получила новое состояние")

			if err = scrap.notifyService.SendUpdates(linkInfo, linkUpdates); err != nil {
				scrap.log.Info("ошибка при отправке обновлений", "err", err.Error())
			}
		}
	}
}
