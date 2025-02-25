package file

import (
	"linkTraccer/internal/domain/scrapper"
	"sync"
)

const (
	userAlreadySaveLink = "пользователь уже сохранял эту ссылку"
	userNotSaveLink     = "пользователь не сохранял эту ссылку"
	userNotRegistered   = "пользователь не регистрировался"
)

type User = scrapper.User
type Link = scrapper.Link
type LinkState = scrapper.LinkState

type FileStorage struct {
	mu          *sync.Mutex
	UserToLinks map[User]map[Link]struct{}
	LinkToUsers map[Link]map[User]struct{}
	LinksState  map[Link]LinkState
}

func NewFileStorage() *FileStorage {
	return &FileStorage{
		mu:          &sync.Mutex{},
		UserToLinks: make(map[User]map[Link]struct{}),
		LinkToUsers: make(map[Link]map[User]struct{}),
		LinksState:  make(map[Link]LinkState),
	}
}

func (f *FileStorage) TrackLink(userID User, userLink Link, initialState LinkState) error {
	f.mu.Lock()

	defer f.mu.Unlock()

	if _, ok := f.UserToLinks[userID]; !ok {
		f.UserToLinks[userID] = make(map[Link]struct{})
	}

	if _, ok := f.LinkToUsers[userLink]; !ok {
		f.LinkToUsers[userLink] = make(map[User]struct{})
	}

	if _, ok := f.UserToLinks[userID][userLink]; ok {
		return NewErrWithStorage(userAlreadySaveLink)
	}

	f.LinksState[userLink] = initialState
	f.UserToLinks[userID][userLink] = struct{}{}
	f.LinkToUsers[userLink][userID] = struct{}{}

	return nil
}

func (f *FileStorage) UntrackLink(user User, link Link) error {
	f.mu.Lock()

	defer f.mu.Unlock()

	if _, ok := f.UserToLinks[user]; !ok {
		return NewErrWithStorage(userNotRegistered)
	}

	if _, ok := f.UserToLinks[user][link]; !ok {
		return NewErrWithStorage(userNotSaveLink)
	}

	if _, ok := f.LinkToUsers[link]; !ok {
		return NewErrWithStorage(userNotSaveLink)
	}

	delete(f.UserToLinks[user], link)
	delete(f.LinkToUsers[link], user)

	if len(f.LinkToUsers[link]) == 0 {
		delete(f.LinkToUsers, link)
		delete(f.LinksState, link)
	}

	return nil
}

func (f *FileStorage) AllUserLinks(user User) ([]Link, error) {
	f.mu.Lock()

	defer f.mu.Unlock()

	if _, ok := f.UserToLinks[user]; !ok || len(f.UserToLinks[user]) == 0 {
		return nil, NewErrWithStorage(userNotSaveLink)
	}

	links := make([]Link, 0, len(f.UserToLinks[user]))

	for userLink, _ := range f.UserToLinks[user] {
		links = append(links, userLink)
	}

	return links, nil
}

func (f *FileStorage) AllLinks() []Link {
	f.mu.Lock()
	defer f.mu.Unlock()

	links := make([]Link, 0, len(f.LinkToUsers))

	for link, _ := range f.LinkToUsers {
		links = append(links, link)
	}

	return links
}

func (f *FileStorage) UsersWhoTrackLink(userLink Link) []User {

	if _, ok := f.LinkToUsers[userLink]; !ok {
		return []User{}
	}

	users := make([]User, 0, len(f.LinkToUsers[userLink]))

	for user, _ := range f.LinkToUsers[userLink] {
		users = append(users, user)
	}

	return users
}

func (f *FileStorage) LinkState(link Link) (LinkState, error) {
	f.mu.Lock()

	defer f.mu.Unlock()

	if _, ok := f.LinksState[link]; !ok {
		return "", NewErrWithStorage(userNotSaveLink)
	}

	return f.LinksState[link], nil
}

func (f *FileStorage) ChangeLinkState(link Link, newState LinkState) error {
	f.mu.Lock()

	defer f.mu.Unlock()

	if _, ok := f.LinksState[link]; !ok {
		return NewErrWithStorage(userNotSaveLink)
	}

	f.LinksState[link] = newState

	return nil
}

func (f *FileStorage) UserExist(userID User) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, ok := f.UserToLinks[userID]

	return ok
}

func (f *FileStorage) DeleteUser(userID User) error {
	if !f.UserExist(userID) {
		return NewErrWithStorage(userNotRegistered)
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	for link, _ := range f.UserToLinks[userID] {

		delete(f.LinkToUsers[link], userID)

		if len(f.LinkToUsers[link]) == 0 {
			delete(f.LinkToUsers, link)
			delete(f.LinksState, link)
		}
	}

	delete(f.UserToLinks, userID)

	return nil
}
