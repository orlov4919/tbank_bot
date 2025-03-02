package scrapperservice

import (
	"linkTraccer/internal/domain/dto"
	"linkTraccer/internal/domain/scrapper"
	"log/slog"
)

type User = scrapper.User
type Link = scrapper.Link
type LinkState = scrapper.LinkState

type UserRepo interface {
	RegUser(userID User) error
	UserTrackLink(user User, link Link) bool
	DeleteUser(userID User) error
	UserExist(userID User) bool
	TrackLink(userID User, userLink Link, initialState LinkState) error
	UntrackLink(user User, link Link) error
	AllUserLinks(user User) ([]Link, error)
	AllLinks() []Link
	UsersWhoTrackLink(userLink Link) []User
	LinkState(link Link) (LinkState, error)
	ChangeLinkState(link Link, newState LinkState) error
}

type SiteClient interface {
	CanTrack(link Link) bool
	LinkState(link Link) (LinkState, error)
}

type BotClient interface {
	SendLinkUpdates(update *dto.LinkUpdate) error
}

type Scrapper struct {
	userRepo    UserRepo
	siteClients []SiteClient
	botClient   BotClient
	log         *slog.Logger
}

const (
	initialLinkState = ""
)

func New(userRepo UserRepo, botClient BotClient, log *slog.Logger, siteClients ...SiteClient) *Scrapper {
	return &Scrapper{
		userRepo:    userRepo,
		botClient:   botClient,
		siteClients: siteClients,
		log:         log,
	}
}

func (scrap *Scrapper) CheckLinkUpdates() {
	for _, link := range scrap.userRepo.AllLinks() {
		for _, siteClient := range scrap.siteClients {
			if !siteClient.CanTrack(link) {
				continue
			}

			linkState, err := siteClient.LinkState(link)

			if err != nil {
				scrap.log.Info("при получении состояния ссылки произошла ошибка", "err", err.Error())
				continue
			}

			savedLinkState, err := scrap.userRepo.LinkState(link)

			if err != nil {
				scrap.log.Info("ошибка при получении состояния ссылки из хранилища", "err", err.Error())
				continue
			}

			if savedLinkState == initialLinkState {
				err := scrap.userRepo.ChangeLinkState(link, linkState)

				if err != nil {
					scrap.log.Info("ошибка при изменении состояния ссылки", "err", err.Error())
				}

				scrap.log.Info("ссылка " + link + " получила начальное состояние")
			} else if linkState != savedLinkState {
				scrap.log.Info("ссылка " + link + " получила новое состояние")

				err := scrap.userRepo.ChangeLinkState(link, linkState)

				if err != nil {
					scrap.log.Info("ошибка при изменении состояния ссылки", "err", err.Error())
					continue
				}

				err = scrap.botClient.SendLinkUpdates(&dto.LinkUpdate{
					ID:          1,
					URL:         link,
					Description: "пришло новое обновление по ссылке",
					TgChatIDs:   scrap.userRepo.UsersWhoTrackLink(link)})

				if err != nil {
					scrap.log.Info("ошибка при отправке обновления ссылки", "err", err.Error())
				}
			}
		}
	}
}
