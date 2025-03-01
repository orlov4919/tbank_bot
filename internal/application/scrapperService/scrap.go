package scrapperService

import (
	"linkTraccer/internal/domain/dto"
	"linkTraccer/internal/domain/scrapper"
	"log"
	"time"
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
}

const (
	initialLinkState = ""
)

func New(userRepo UserRepo, botClient BotClient, siteClients ...SiteClient) *Scrapper {
	return &Scrapper{
		userRepo:    userRepo,
		botClient:   botClient,
		siteClients: siteClients,
	}
}

func (scrap *Scrapper) CheckLinkUpdates() {

	for {

		for _, link := range scrap.userRepo.AllLinks() {

			for _, siteClient := range scrap.siteClients {

				if !siteClient.CanTrack(link) {
					continue
				}

				linkState, err := siteClient.LinkState(link)

				if err != nil {
					log.Println("При получении состояни ссылки от клиента, произошла ошибка")
					continue
				}

				savedLinkState, err := scrap.userRepo.LinkState(link)

				if err != nil {
					log.Println("При получении состояния ссылки из хранилища, произошла ошибка")
					continue
				}

				if savedLinkState == initialLinkState {
					err := scrap.userRepo.ChangeLinkState(link, linkState)

					if err != nil {
						log.Println("Ошибка при изменении состояния ссылки")
					}

				} else if linkState != savedLinkState {

					err := scrap.userRepo.ChangeLinkState(link, linkState)

					if err != nil {
						log.Println("Ошибка при изменении состояния ссылки")
					} else {
						log.Println("Ссылка получила новое состояние")

						scrap.botClient.SendLinkUpdates(&dto.LinkUpdate{
							ID:          1,
							URL:         link,
							Description: "пришло новое обновление по ссылке",
							TgChatIds:   scrap.userRepo.UsersWhoTrackLink(link),
						})
					}
				}
			}
		}

		time.Sleep(time.Minute)
	}
}
