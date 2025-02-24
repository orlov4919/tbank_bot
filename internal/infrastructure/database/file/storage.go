package file

import (
	"errors"
	"linkTraccer/internal/domain/scrapper"
	"sync"
)

type User = scrapper.User
type Link = scrapper.Link
type LinkState = scrapper.LinkState

type FileStorage struct {
	mu          *sync.Mutex
	userToLinks map[User]map[Link]struct{}
	linkToUsers map[Link]map[User]struct{}
	linkState   map[Link]LinkState
}

func NewFileStorage() *FileStorage {
	return &FileStorage{
		mu:          &sync.Mutex{},
		userToLinks: make(map[User]map[Link]struct{}),
		linkToUsers: make(map[Link]map[User]struct{}),
		linkState:   make(map[Link]LinkState),
	}
}

func (f *FileStorage) TrackLink(userID User, userLink Link) error {
	f.mu.Lock()

	defer f.mu.Unlock()

	if _, ok := f.userToLinks[userID]; !ok {
		f.userToLinks[userID] = make(map[Link]struct{})
	}

	if _, ok := f.linkToUsers[userLink]; !ok {
		f.linkToUsers[userLink] = make(map[User]struct{})
	}

	if _, ok := f.userToLinks[userID][userLink]; ok {
		return errors.New("Ссылка уже сохранена")
	}

	f.linkState[userLink] = ""
	f.userToLinks[userID][userLink] = struct{}{}
	f.linkToUsers[userLink][userID] = struct{}{}

	return nil
}

func (f *FileStorage) UntrackLink(user User, link Link) error {
	f.mu.Lock()

	defer f.mu.Unlock()

	if _, ok := f.userToLinks[user]; !ok {
		return errors.New("нет указанного пользователя")
	}

	if _, ok := f.userToLinks[user][link]; !ok {
		return errors.New("указанный пользователь не отслеживает эту ссылку")
	}

	delete(f.userToLinks[user], link)
	delete(f.linkToUsers[link], user) // вроде бы ошибка

	if len(f.linkToUsers[link]) == 0 {
		delete(f.linkToUsers, link)
	}

	return nil
}

func (f *FileStorage) AllUserLinks(user User) ([]Link, error) {
	f.mu.Lock()

	defer f.mu.Unlock()

	if _, ok := f.userToLinks[user]; !ok || len(f.userToLinks[user]) == 0 {
		return nil, errors.New("нет указанной ссылки в хранилище")
	}

	links := make([]Link, 0, len(f.userToLinks[user]))

	for userLink, _ := range f.userToLinks[user] {
		links = append(links, userLink)
	}

	return links, nil
}

func (f *FileStorage) AllLinks() []Link {
	f.mu.Lock()
	defer f.mu.Unlock()

	links := make([]Link, 0, len(f.linkToUsers))

	for link, _ := range f.linkToUsers {
		links = append(links, link)
	}

	return links
}

func (f *FileStorage) UsersWhoTrackLink(userLink Link) []User {

	if _, ok := f.linkToUsers[userLink]; !ok {
		return []User{}
	}

	users := make([]User, 0, len(f.linkToUsers[userLink]))

	for user, _ := range f.linkToUsers[userLink] {
		users = append(users, user)
	}

	return users
}

func (f *FileStorage) LinkState(link Link) (LinkState, error) {
	f.mu.Lock()

	defer f.mu.Unlock()

	if _, ok := f.linkState[link]; !ok {
		return "", errors.New("Нет такой ссылки в хранилище")
	}

	return f.linkState[link], nil

}

func (f *FileStorage) ChangeLinkState(link Link, newState LinkState) error {
	f.mu.Lock()

	defer f.mu.Unlock()

	if _, ok := f.linkState[link]; !ok {
		return errors.New("Нет такой ссылки в хранилище")
	}

	f.linkState[link] = newState

	return nil
}
