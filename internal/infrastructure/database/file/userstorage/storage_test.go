package userstorage_test

import (
	"linkTraccer/internal/domain/scrapper"
	"linkTraccer/internal/infrastructure/database/file/userstorage"
	"testing"

	"github.com/stretchr/testify/assert"
)

type User = scrapper.User
type Link = scrapper.Link
type LinkState = scrapper.LinkState

const (
	zeroState = ""
)

func TestFileStorage_RegUser(t *testing.T) {
	type testCase struct {
		name    string
		id      User
		correct bool
	}

	tests := []testCase{
		{
			name:    "Первый раз добавляем юзера в хранилище",
			id:      1,
			correct: true,
		},
		{
			name:    "Повторно пытаемся добавить юзера в хранилище",
			id:      1,
			correct: false,
		},
	}

	store := userstorage.NewFileStorage()

	for _, test := range tests {
		err := store.RegUser(test.id)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestFileStorage_TrackLink(t *testing.T) {
	type testCase struct {
		name      string
		userID    User
		userLink  Link
		linkState LinkState

		userToLinks map[User]map[Link]struct{}
		linkToUsers map[Link]map[User]struct{}
		linksState  map[Link]LinkState

		correct bool
	}

	tests := []testCase{
		{
			name:      "Простой тест на добавление ссылки (добавление нового пользователя)",
			userID:    1,
			userLink:  "google.com",
			linkState: zeroState,

			userToLinks: map[User]map[Link]struct{}{
				1: {"google.com": struct{}{}},
			},
			linkToUsers: map[Link]map[User]struct{}{
				"google.com": {1: struct{}{}},
			},
			linksState: map[Link]LinkState{
				"google.com": zeroState,
			},

			correct: true,
		},
		{
			name:      "Тест на добавление повторной ссылки, должен вызвать ошибку",
			userID:    1,
			userLink:  "google.com",
			linkState: zeroState,

			userToLinks: map[User]map[Link]struct{}{
				1: {"google.com": struct{}{}},
			},
			linkToUsers: map[Link]map[User]struct{}{
				"google.com": {1: struct{}{}},
			},
			linksState: map[Link]LinkState{
				"google.com": zeroState,
			},

			correct: false,
		},
		{
			name:      "Тест на добавление новой ссылки, при том что пользователь уже сохранен",
			userID:    1,
			userLink:  "tbank.ru",
			linkState: zeroState,

			userToLinks: map[User]map[Link]struct{}{
				1: {
					"google.com": struct{}{},
					"tbank.ru":   struct{}{}},
			},
			linkToUsers: map[Link]map[User]struct{}{
				"google.com": {1: struct{}{}},
				"tbank.ru":   {1: struct{}{}},
			},
			linksState: map[Link]LinkState{
				"google.com": zeroState,
				"tbank.ru":   zeroState,
			},

			correct: true,
		},
		{
			name:     "Тест на добавление нового пользователя, для уже имеющейся в базе ссылки",
			userID:   2,
			userLink: "tbank.ru",

			userToLinks: map[User]map[Link]struct{}{
				1: {
					"google.com": struct{}{},
					"tbank.ru":   struct{}{}},
				2: {"tbank.ru": struct{}{}},
			},
			linkToUsers: map[Link]map[User]struct{}{
				"google.com": {1: struct{}{}},
				"tbank.ru":   {1: struct{}{}, 2: struct{}{}},
			},
			linksState: map[Link]LinkState{
				"google.com": zeroState,
				"tbank.ru":   zeroState,
			},

			linkState: zeroState,
			correct:   true,
		},
	}

	fileStorage := userstorage.NewFileStorage()

	for _, test := range tests {
		_ = fileStorage.RegUser(test.userID)

		err := fileStorage.TrackLink(test.userID, test.userLink, test.linkState)

		assert.Equal(t, fileStorage.UserToLinks, test.userToLinks)
		assert.Equal(t, fileStorage.LinkToUsers, test.linkToUsers)
		assert.Equal(t, fileStorage.LinksState, test.linksState)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestFileStorage_UntrackLink(t *testing.T) {
	// для начала добавим данные в хранилище
	type dataToStore struct {
		name      string
		userID    User
		userLink  Link
		linkState LinkState
	}

	dataSlice := []dataToStore{
		{
			name:      "Добавляем первую ссылку у юзера с id = 1",
			userID:    1,
			userLink:  "google.com",
			linkState: zeroState,
		},
		{
			name:      "Добавляем вторую ссылку у юзера с id = 1",
			userID:    1,
			userLink:  "tbank.ru",
			linkState: zeroState,
		},
		{
			name:      "Добавляем первую ссылку у юзера с id = 2",
			userID:    2,
			userLink:  "tbank.ru",
			linkState: zeroState,
		},
	}

	fileStorage := userstorage.NewFileStorage()

	for _, data := range dataSlice {
		_ = fileStorage.RegUser(data.userID)
		_ = fileStorage.TrackLink(data.userID, data.userLink, data.linkState)
	}

	type testCase struct {
		name string

		userID   User
		userLink Link

		userToLinks map[User]map[Link]struct{}
		linkToUsers map[Link]map[User]struct{}
		linksState  map[Link]LinkState

		correct bool
	}

	tests := []testCase{
		{
			name:     "Тест на удаление ссылки незарегистрированного пользователя",
			userID:   3,
			userLink: "google.com",

			userToLinks: map[User]map[Link]struct{}{
				1: {
					"google.com": struct{}{},
					"tbank.ru":   struct{}{}},
				2: {"tbank.ru": struct{}{}},
			},
			linkToUsers: map[Link]map[User]struct{}{
				"google.com": {1: struct{}{}},
				"tbank.ru":   {1: struct{}{}, 2: struct{}{}},
			},
			linksState: map[Link]LinkState{
				"google.com": zeroState,
				"tbank.ru":   zeroState,
			},

			correct: false,
		},
		{
			name:     "Тест на удаление ссылки, которую пользователь не отслеживает",
			userID:   2,
			userLink: "google.com",

			userToLinks: map[User]map[Link]struct{}{
				1: {
					"google.com": struct{}{},
					"tbank.ru":   struct{}{}},
				2: {"tbank.ru": struct{}{}},
			},
			linkToUsers: map[Link]map[User]struct{}{
				"google.com": {1: struct{}{}},
				"tbank.ru":   {1: struct{}{}, 2: struct{}{}},
			},
			linksState: map[Link]LinkState{
				"google.com": zeroState,
				"tbank.ru":   zeroState,
			},

			correct: false,
		},
		{
			name:     "Тест на удаление ссылки, которую пользователь отслеживает",
			userID:   2,
			userLink: "tbank.ru",

			userToLinks: map[User]map[Link]struct{}{
				1: {
					"google.com": struct{}{},
					"tbank.ru":   struct{}{}},
				2: {},
			},
			linkToUsers: map[Link]map[User]struct{}{
				"google.com": {1: struct{}{}},
				"tbank.ru":   {1: struct{}{}},
			},
			linksState: map[Link]LinkState{
				"google.com": zeroState,
				"tbank.ru":   zeroState,
			},

			correct: true,
		},
		{
			name:     "Тест на удаление ссылки, которую пользователь отслеживает",
			userID:   1,
			userLink: "google.com",

			userToLinks: map[User]map[Link]struct{}{
				1: {"tbank.ru": struct{}{}},
				2: {},
			},
			linkToUsers: map[Link]map[User]struct{}{
				"tbank.ru": {1: struct{}{}},
			},
			linksState: map[Link]LinkState{
				"tbank.ru": zeroState,
			},

			correct: true,
		},
	}

	for _, test := range tests {
		err := fileStorage.UntrackLink(test.userID, test.userLink)

		assert.Equal(t, test.userToLinks, fileStorage.UserToLinks)
		assert.Equal(t, test.linkToUsers, fileStorage.LinkToUsers)
		assert.Equal(t, test.linksState, fileStorage.LinksState)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestFileStorage_GetAllUserLinks(t *testing.T) {
	// для начала добавим данные в хранилище
	type dataToStore struct {
		name      string
		userID    User
		userLink  Link
		linkState LinkState
	}

	dataSlice := []dataToStore{
		{
			name:      "Добавляем первую ссылку у юзера с id = 1",
			userID:    1,
			userLink:  "google.com",
			linkState: zeroState,
		},
		{
			name:      "Добавляем вторую ссылку у юзера с id = 1",
			userID:    1,
			userLink:  "tbank.ru",
			linkState: zeroState,
		},
		{
			name:      "Добавляем первую ссылку у юзера с id = 2",
			userID:    2,
			userLink:  "etu.ru",
			linkState: zeroState,
		},
	}

	fileStorage := userstorage.NewFileStorage()

	for _, data := range dataSlice {
		_ = fileStorage.RegUser(data.userID)
		_ = fileStorage.TrackLink(data.userID, data.userLink, data.linkState)
	}

	type testCase struct {
		name    string
		userID  User
		links   []Link
		correct bool
	}

	tests := []testCase{
		{
			name:    "Проверим все ссылки первого пользователя",
			userID:  1,
			links:   []Link{"google.com", "tbank.ru"},
			correct: true,
		},
		{
			name:    "Проверим все ссылки второго пользователя",
			userID:  2,
			links:   []Link{"etu.ru"},
			correct: true,
		},
		{
			name:    "Проверим все ссылки третьего пользователя(ожидаем ошибку)",
			userID:  3,
			links:   nil,
			correct: false,
		},
	}

	for _, test := range tests {
		links, err := fileStorage.AllUserLinks(test.userID)

		if test.correct {
			assert.ElementsMatch(t, test.links, links)
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestFileStorage_GetUsersWhoTrackLink(t *testing.T) {
	type dataToStore struct {
		name      string
		userID    User
		userLink  Link
		linkState LinkState
	}

	dataSlice := []dataToStore{
		{
			name:      "Добавляем первую ссылку у юзера с id = 1",
			userID:    1,
			userLink:  "google.com",
			linkState: zeroState,
		},
		{
			name:      "Добавляем вторую ссылку у юзера с id = 1",
			userID:    1,
			userLink:  "tbank.ru",
			linkState: zeroState,
		},
		{
			name:      "Добавляем первую ссылку у юзера с id = 2",
			userID:    2,
			userLink:  "etu.ru",
			linkState: zeroState,
		},
		{
			name:      "Добавляем вторую ссылку для юзера с id = 2",
			userID:    2,
			userLink:  "tbank.ru",
			linkState: zeroState,
		},
	}

	fileStorage := userstorage.NewFileStorage()

	for _, data := range dataSlice {
		_ = fileStorage.RegUser(data.userID)
		_ = fileStorage.TrackLink(data.userID, data.userLink, data.linkState)
	}

	type testCase struct {
		name  string
		link  Link
		users []User
	}

	tests := []testCase{
		{
			name:  "Проверяем кто следит за ссылкой tbank.ru",
			link:  "tbank.ru",
			users: []User{1, 2},
		},
		{
			name:  "Проверяем кто следит за ссылкой google.com",
			link:  "google.com",
			users: []User{1},
		},
		{
			name:  "Проверяем кто следит за ссылкой etu.ru",
			link:  "etu.ru",
			users: []User{2},
		},
		{
			name:  "Проверяем кто следит за ссылкой alo.ru (ожидаем что никто)",
			link:  "alo.ru",
			users: []User{},
		},
	}

	for _, test := range tests {
		assert.ElementsMatch(t, test.users, fileStorage.UsersWhoTrackLink(test.link))
	}
}

func TestFileStorage_AllLinks(t *testing.T) {
	type dataToStore struct {
		name      string
		userID    User
		userLink  Link
		linkState LinkState
	}

	dataSlice := []dataToStore{
		{
			name:      "Добавляем первую ссылку у юзера с id = 1",
			userID:    1,
			userLink:  "google.com",
			linkState: zeroState,
		},
		{
			name:      "Добавляем вторую ссылку у юзера с id = 1",
			userID:    1,
			userLink:  "tbank.ru",
			linkState: zeroState,
		},
		{
			name:      "Добавляем первую ссылку у юзера с id = 2",
			userID:    2,
			userLink:  "etu.ru",
			linkState: zeroState,
		},
		{
			name:      "Добавляем вторую ссылку для юзера с id = 2",
			userID:    2,
			userLink:  "tbank.ru",
			linkState: zeroState,
		},
	}

	fileStorage := userstorage.NewFileStorage()

	for _, data := range dataSlice {
		_ = fileStorage.RegUser(data.userID)
		_ = fileStorage.TrackLink(data.userID, data.userLink, data.linkState)
	}

	type testCase struct {
		name  string
		links []Link
	}

	tests := []testCase{
		{
			name:  "Проверяем, что все ссылки успешно сохранены",
			links: []Link{"tbank.ru", "etu.ru", "google.com"},
		},
	}

	for _, test := range tests {
		assert.ElementsMatch(t, test.links, fileStorage.AllLinks())
	}
}

func TestFileStorage_LinkState(t *testing.T) {
	type dataToStore struct {
		name      string
		userID    User
		userLink  Link
		linkState LinkState
	}

	dataSlice := []dataToStore{
		{
			name:      "Добавляем первую ссылку у юзера с id = 1",
			userID:    1,
			userLink:  "google.com",
			linkState: zeroState,
		},
		{
			name:      "Добавляем вторую ссылку у юзера с id = 1",
			userID:    1,
			userLink:  "tbank.ru",
			linkState: zeroState,
		},
		{
			name:      "Добавляем первую ссылку у юзера с id = 2",
			userID:    2,
			userLink:  "etu.ru",
			linkState: zeroState,
		},
		{
			name:      "Добавляем вторую ссылку для юзера с id = 2",
			userID:    2,
			userLink:  "tbank.ru",
			linkState: zeroState,
		},
	}

	fileStorage := userstorage.NewFileStorage()

	for _, data := range dataSlice {
		_ = fileStorage.RegUser(data.userID)
		_ = fileStorage.TrackLink(data.userID, data.userLink, data.linkState)
	}

	type testCase struct {
		name      string
		link      Link
		LinkState LinkState
		correct   bool
	}

	tests := []testCase{
		{
			name:      "Проверяем состояние ссылки tbank.ru",
			link:      "tbank.ru",
			LinkState: zeroState,
			correct:   true,
		},
		{
			name:    "Проверяем состояние ссылки, которую не добавляли",
			link:    "mak.ru",
			correct: false,
		},
	}

	for _, test := range tests {
		resState, err := fileStorage.LinkState(test.link)

		if test.correct {
			assert.NoError(t, err)
			assert.Equal(t, test.LinkState, resState)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestFileStorage_ChangeLinkState(t *testing.T) {
	type dataToStore struct {
		name      string
		userID    User
		userLink  Link
		linkState LinkState
	}

	dataSlice := []dataToStore{
		{
			name:      "Добавляем первую ссылку у юзера с id = 1",
			userID:    1,
			userLink:  "google.com",
			linkState: zeroState,
		},
		{
			name:      "Добавляем вторую ссылку у юзера с id = 1",
			userID:    1,
			userLink:  "tbank.ru",
			linkState: zeroState,
		},
		{
			name:      "Добавляем первую ссылку у юзера с id = 2",
			userID:    2,
			userLink:  "etu.ru",
			linkState: zeroState,
		},
		{
			name:      "Добавляем вторую ссылку для юзера с id = 2",
			userID:    2,
			userLink:  "tbank.ru",
			linkState: zeroState,
		},
	}

	fileStorage := userstorage.NewFileStorage()

	for _, data := range dataSlice {
		_ = fileStorage.RegUser(data.userID)
		_ = fileStorage.TrackLink(data.userID, data.userLink, data.linkState)
	}

	type testCase struct {
		name     string
		link     Link
		newState LinkState
		correct  bool
	}

	tests := []testCase{
		{
			name:     "Пробуем изменить состояние не добавленной ссылки",
			link:     "mak.by",
			newState: "state2",
			correct:  false,
		},
		{
			name:     "Пробуем изменить состояние добавленной ссылки",
			link:     "tbank.ru",
			newState: "state2",
			correct:  true,
		},
	}

	for _, test := range tests {
		err := fileStorage.ChangeLinkState(test.link, test.newState)

		if !test.correct {
			assert.Error(t, err)
		} else {
			linkStateInStorage, _ := fileStorage.LinkState(test.link)

			assert.NoError(t, err)
			assert.Equal(t, test.newState, linkStateInStorage)
		}
	}
}

func TestFileStorage_UserExist(t *testing.T) {
	type dataToStore struct {
		name      string
		userID    User
		userLink  Link
		linkState LinkState
	}

	dataSlice := []dataToStore{
		{
			name:      "Добавляем первую ссылку у юзера с id = 1",
			userID:    1,
			userLink:  "google.com",
			linkState: zeroState,
		},
		{
			name:      "Добавляем вторую ссылку у юзера с id = 1",
			userID:    1,
			userLink:  "tbank.ru",
			linkState: zeroState,
		},
		{
			name:      "Добавляем первую ссылку у юзера с id = 2",
			userID:    2,
			userLink:  "etu.ru",
			linkState: zeroState,
		},
		{
			name:      "Добавляем вторую ссылку для юзера с id = 2",
			userID:    2,
			userLink:  "tbank.ru",
			linkState: zeroState,
		},
	}

	fileStorage := userstorage.NewFileStorage()

	for _, data := range dataSlice {
		_ = fileStorage.RegUser(data.userID)
		_ = fileStorage.TrackLink(data.userID, data.userLink, data.linkState)
	}

	type testCase struct {
		name      string
		user      User
		userExist bool
	}

	tests := []testCase{
		{
			name:      "Проверяем содержится ли юзер, которого мы не добавляли",
			user:      20,
			userExist: false,
		},
		{
			name:      "Проверяем содержится ли юзер, которого мы добавляли",
			user:      2,
			userExist: true,
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.userExist, fileStorage.UserExist(test.user))
	}
}

func TestFileStorage_DeleteUser(t *testing.T) {
	type dataToStore struct {
		name      string
		userID    User
		userLink  Link
		linkState LinkState
	}

	dataSlice := []dataToStore{
		{
			name:      "Добавляем первую ссылку у юзера с id = 1",
			userID:    1,
			userLink:  "google.com",
			linkState: zeroState,
		},
		{
			name:      "Добавляем вторую ссылку у юзера с id = 1",
			userID:    1,
			userLink:  "tbank.ru",
			linkState: zeroState,
		},
		{
			name:      "Добавляем первую ссылку у юзера с id = 2",
			userID:    2,
			userLink:  "tbank.ru",
			linkState: zeroState,
		},
	}

	fileStorage := userstorage.NewFileStorage()

	for _, data := range dataSlice {
		_ = fileStorage.RegUser(data.userID)
		_ = fileStorage.TrackLink(data.userID, data.userLink, data.linkState)
	}

	type testCase struct {
		name string

		userID User

		userToLinks map[User]map[Link]struct{}
		linkToUsers map[Link]map[User]struct{}
		linksState  map[Link]LinkState

		correct bool
	}

	tests := []testCase{
		{
			name:   "Тест на удаление незарегистрированного пользователя",
			userID: 3,

			userToLinks: map[User]map[Link]struct{}{
				1: {
					"google.com": struct{}{},
					"tbank.ru":   struct{}{}},
				2: {"tbank.ru": struct{}{}},
			},
			linkToUsers: map[Link]map[User]struct{}{
				"google.com": {1: struct{}{}},
				"tbank.ru":   {1: struct{}{}, 2: struct{}{}},
			},
			linksState: map[Link]LinkState{
				"google.com": zeroState,
				"tbank.ru":   zeroState,
			},

			correct: false,
		},

		{
			name:   "Тест на удаление пользователя  2",
			userID: 2,

			userToLinks: map[User]map[Link]struct{}{
				1: {
					"google.com": struct{}{},
					"tbank.ru":   struct{}{}},
			},
			linkToUsers: map[Link]map[User]struct{}{
				"google.com": {1: struct{}{}},
				"tbank.ru":   {1: struct{}{}},
			},
			linksState: map[Link]LinkState{
				"google.com": zeroState,
				"tbank.ru":   zeroState,
			},

			correct: true,
		},
		{
			name:   "Тест на удаление пользователя  1",
			userID: 1,

			userToLinks: map[User]map[Link]struct{}{},
			linkToUsers: map[Link]map[User]struct{}{},
			linksState:  map[Link]LinkState{},

			correct: true,
		},
	}

	for _, test := range tests {
		err := fileStorage.DeleteUser(test.userID)

		assert.Equal(t, test.userToLinks, fileStorage.UserToLinks)
		assert.Equal(t, test.linkToUsers, fileStorage.LinkToUsers)
		assert.Equal(t, test.linksState, fileStorage.LinksState)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestFileStorage_UserTrackLink(t *testing.T) {
	type dataToStore struct {
		name      string
		userID    User
		userLink  Link
		linkState LinkState
	}

	dataSlice := []dataToStore{
		{
			name:      "Добавляем первую ссылку у юзера с id = 1",
			userID:    1,
			userLink:  "google.com",
			linkState: zeroState,
		},
	}

	fileStorage := userstorage.NewFileStorage()

	for _, data := range dataSlice {
		_ = fileStorage.RegUser(data.userID)
		_ = fileStorage.TrackLink(data.userID, data.userLink, data.linkState)
	}

	type testCase struct {
		name string

		userID   User
		userLink Link

		correct bool
	}

	tests := []testCase{
		{
			name:     "Тест на пользователя который не зарегистрирован",
			userID:   3,
			userLink: "google.com",
			correct:  false,
		},
		{
			name:     "Тест на ссылку, которую пользователь не отслеживает",
			userID:   1,
			userLink: "tbank.ru",
			correct:  false,
		},
		{
			name:     "Тест на ссылку, которую пользователь отслеживает",
			userID:   1,
			userLink: "google.com",
			correct:  true,
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.correct, fileStorage.UserTrackLink(test.userID, test.userLink))
	}
}
