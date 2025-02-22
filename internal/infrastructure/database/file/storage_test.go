package file_test

import (
	"github.com/stretchr/testify/assert"
	"linkTraccer/internal/domain/scrapper"
	"linkTraccer/internal/infrastructure/database/file"
	"testing"
)

func TestFileStorage_TrackLink(t *testing.T) {
	type testCase struct {
		name     string
		userID   int
		userLink string
		correct  bool
	}

	tests := []testCase{
		{
			name:     "Простой тест на добавление ссылки (добавление нового пользователя)",
			userID:   1,
			userLink: "google.com",
			correct:  true,
		},
		{
			name:     "Тест на добавление повторной ссылки, должен вызвать ошибку",
			userID:   1,
			userLink: "google.com",
			correct:  false,
		},
		{
			name:     "Тест на добавление новой ссылки, при том что пользователь уже сохранен",
			userID:   1,
			userLink: "tbank.com",
			correct:  true,
		},
		{
			name:     "Тест на добавление нового пользователя, для уже имеющейся в базе ссылки",
			userID:   2,
			userLink: "tbank.com",
			correct:  true,
		},
	}

	fileStorage := file.NewFileStorage()

	for _, test := range tests {
		err := fileStorage.TrackLink(test.userID, test.userLink)

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
		name     string
		userID   int
		userLink string
	}

	dataSlice := []dataToStore{
		{
			name:     "Добавляем первую ссылку у юзера с id = 1",
			userID:   1,
			userLink: "google.com",
		},
		{
			name:     "Добавляем вторую ссылку у юзера с id = 1",
			userID:   1,
			userLink: "tbank.ru",
		},
		{
			name:     "Добавляем первую ссылку у юзера с id = 2",
			userID:   2,
			userLink: "google.com",
		},
		{
			name:     "Добавляем вторую ссылку у юзера с id = 2",
			userID:   2,
			userLink: "tbank.ru",
		},
		{
			name:     "Добавляем первую ссылку у юзера с id = 3",
			userID:   3,
			userLink: "etu.ru",
		},
	}

	fileStorage := file.NewFileStorage()

	for _, data := range dataSlice {
		fileStorage.TrackLink(data.userID, data.userLink)
	}

	type testCase struct {
		name     string
		userID   int
		userLink string
		correct  bool
	}

	tests := []testCase{
		{
			name:     "Удалим у пользователя 3, единственную хранимую ссылку",
			userID:   3,
			userLink: "etu.ru",
			correct:  true,
		},
		{
			name:     "Попробуем повторно удалить ссылку пользователя 3 (ожидаем ошибку)",
			userID:   3,
			userLink: "etu.ru",
			correct:  false,
		},
		{
			name:     "Удалим у пользователя 2, ссылку на google",
			userID:   2,
			userLink: "google.com",
			correct:  true,
		},
		{
			name:     "Удалим у пользователя 1, ссылку на google",
			userID:   1,
			userLink: "google.com",
			correct:  true,
		},
		{
			name:     "Удалим у пользователя 2, повторно ссылку на google (ожидаем ошибку)",
			userID:   2,
			userLink: "google.com",
			correct:  false,
		},
		{
			name:     "Попробуем удалить у пользователя 3, ссылку на tbank (ожидаем ошибку)",
			userID:   3,
			userLink: "tbank.ru",
			correct:  false,
		},
		{
			name:     "Попробуем удалить у пользователя 4, ссылку на tbank (ожидаем ошибку)",
			userID:   4,
			userLink: "tbank.ru",
			correct:  false,
		},
	}

	for _, test := range tests {

		err := fileStorage.UntrackLink(test.userID, test.userLink)

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
		name     string
		userID   int
		userLink string
	}

	dataSlice := []dataToStore{
		{
			name:     "Добавляем первую ссылку у юзера с id = 1",
			userID:   1,
			userLink: "google.com",
		},
		{
			name:     "Добавляем вторую ссылку у юзера с id = 1",
			userID:   1,
			userLink: "tbank.ru",
		},
		{
			name:     "Добавляем первую ссылку у юзера с id = 2",
			userID:   2,
			userLink: "etu.ru",
		},
	}

	fileStorage := file.NewFileStorage()

	for _, data := range dataSlice {
		fileStorage.TrackLink(data.userID, data.userLink)
	}

	type testCase struct {
		name    string
		userID  int
		links   []scrapper.Link
		correct bool
	}

	tests := []testCase{
		{
			name:    "Проверим все ссылки первого пользователя",
			userID:  1,
			links:   []scrapper.Link{"google.com", "tbank.ru"},
			correct: true,
		},
		{
			name:    "Проверим все ссылки второго пользователя",
			userID:  2,
			links:   []scrapper.Link{"etu.ru"},
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

		links, err := fileStorage.GetAllUserLinks(test.userID)

		if test.correct {
			assert.ElementsMatch(t, test.links, links)
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}

	}
}

func TestFileStorage_GetAllLinks(t *testing.T) {

	type dataToStore struct {
		name     string
		userID   int
		userLink string
	}

	dataSlice := []dataToStore{
		{
			name:     "Добавляем первую ссылку у юзера с id = 1",
			userID:   1,
			userLink: "google.com",
		},
		{
			name:     "Добавляем вторую ссылку у юзера с id = 1",
			userID:   1,
			userLink: "tbank.ru",
		},
		{
			name:     "Добавляем первую ссылку у юзера с id = 2",
			userID:   2,
			userLink: "etu.ru",
		},
	}

	fileStorage := file.NewFileStorage()

	for _, data := range dataSlice {
		fileStorage.TrackLink(data.userID, data.userLink)
	}

	type testCase struct {
		name  string
		links []scrapper.Link
	}

	tests := []testCase{
		{
			name:  "Проверяем, что вернуться все добавленнные ссылки",
			links: []scrapper.Link{"google.com", "tbank.ru", "etu.ru"},
		},
	}

	for _, test := range tests {
		assert.ElementsMatch(t, test.links, fileStorage.GetAllLinks())
	}
}

func TestFileStorage_GetUsersWhoTrackLink(t *testing.T) {

	type dataToStore struct {
		name     string
		userID   int
		userLink string
	}

	dataSlice := []dataToStore{
		{
			name:     "Добавляем первую ссылку у юзера с id = 1",
			userID:   1,
			userLink: "google.com",
		},
		{
			name:     "Добавляем вторую ссылку у юзера с id = 1",
			userID:   1,
			userLink: "tbank.ru",
		},
		{
			name:     "Добавляем первую ссылку у юзера с id = 2",
			userID:   2,
			userLink: "etu.ru",
		},
		{
			name:     "Добавляем вторую ссылку для юзера с id = 2",
			userID:   2,
			userLink: "tbank.ru",
		},
	}

	fileStorage := file.NewFileStorage()

	for _, data := range dataSlice {
		fileStorage.TrackLink(data.userID, data.userLink)
	}

	type testCase struct {
		name  string
		link  scrapper.Link
		users []scrapper.User
	}

	tests := []testCase{
		{
			name:  "Проверяем кто следит за ссылкой tbank.ru",
			link:  "tbank.ru",
			users: []scrapper.User{1, 2},
		},
		{
			name:  "Проверяем кто следит за ссылкой google.com",
			link:  "google.com",
			users: []scrapper.User{1},
		},
		{
			name:  "Проверяем кто следит за ссылкой etu.ru",
			link:  "etu.ru",
			users: []scrapper.User{2},
		},
		{
			name:  "Проверяем кто следит за ссылкой alo.ru (ожидаем что никто)",
			link:  "alo.ru",
			users: []scrapper.User{},
		},
	}

	for _, test := range tests {
		assert.ElementsMatch(t, test.users, fileStorage.GetUsersWhoTrackLink(test.link))
	}
}
