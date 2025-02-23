package scrapperService

import "linkTraccer/internal/domain/scrapper"

type User = scrapper.User
type Link = scrapper.Link
type LinkState = scrapper.LinkState

type UserRepo interface {
	TrackLink(userID User, userLink Link) error
	UntrackLink(user User, link Link) error
	GetAllUserLinks(user User) ([]Link, error)
	GetAllLinks() []Link
	GetUsersWhoTrackLink(userLink Link) []User
}

type Client interface {
	CanTrack(link Link) bool
	LinkState(link Link) (LinkState, error)
}

type BotClient interface {
}

type Scrapper struct {
	userRepo UserRepo
	clients  []Client
}
