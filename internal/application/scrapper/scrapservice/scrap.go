package scrapservice

import (
	"context"
	"linkTraccer/internal/domain/scrapper"
	"log/slog"
	"time"
)

const (
	descriptionFormat = "Пришло новое уведомление 🔥\n\nСобытие: %s\nПользователь: %s\nВремя создаения: %s\nПревью: %s\n\n"
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
	LinksBatch() ([]LinkInfo, error)
	//HasNext() bool // для удобства был бы крут, но так хыз
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
	SendUpdates(linkInfo LinkInfo, linkUpdates LinkUpdates) error
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
	links, err := linksPaginator.LinksBatch()

	for len(links) != 0 && err == nil {
		for _, linkInfo := range links {
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

		links, err = linksPaginator.LinksBatch()
	}

	if err != nil {
		scrap.log.Info("ошибка при проверке ссылок", "err", err.Error())
	}
}
