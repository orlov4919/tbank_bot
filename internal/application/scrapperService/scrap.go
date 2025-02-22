package scrapperService

import "linkTraccer/internal/domain/scrapper"

type User = scrapper.User
type Link = scrapper.Link

type UserRepo interface {
	TrackLink(userID User, userLink Link) error
	UntrackLink(user User, link Link) error
	GetAllUserLinks(user User) ([]Link, error)
	GetAllLinks() []Link
	GetUsersWhoTrackLink(userLink Link) []User
}
