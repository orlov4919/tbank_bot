package contextStorage_test

import (
	"github.com/stretchr/testify/assert"
	"linkTraccer/internal/domain/tgbot"
	"linkTraccer/internal/infrastructure/database/file/contextStorage"
	"testing"
)

type ID = tgbot.ID
type ContextData = tgbot.ContextData

func TestContextStorage_RegUser(t *testing.T) {

	type testCase struct {
		name    string
		id      ID
		correct bool
	}

	tests := []testCase{
		{
			name:    "Производим добавление пользователя",
			id:      1,
			correct: true,
		},
		{
			name:    "Производим повторное добавление пользователя",
			id:      1,
			correct: false,
		},
	}

	contextStore := contextStorage.New()

	for _, test := range tests {
		err := contextStore.RegUser(test.id)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestContextStorage_AddUrl(t *testing.T) {

	contextStore := contextStorage.New()

	type StoreData struct {
		id ID
	}

	data := []StoreData{
		{id: 1},
	}

	for _, user := range data {
		contextStore.RegUser(user.id)
	}

	type testCase struct {
		name    string
		id      ID
		URL     string
		correct bool
	}

	tests := []testCase{
		{
			name:    "Производим добавление ccылки пользователя",
			id:      1,
			URL:     "google.com",
			correct: true,
		},
		{
			name:    "Производим добавление ссылки не зарегистрированного пользователя",
			id:      2,
			URL:     "tbank.ru",
			correct: false,
		},
	}

	for _, test := range tests {
		err := contextStore.AddUrl(test.id, test.URL)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestContextStorage_AddFilters(t *testing.T) {

	contextStore := contextStorage.New()

	type StoreData struct {
		id ID
	}

	data := []StoreData{
		{id: 1},
	}

	for _, user := range data {
		contextStore.RegUser(user.id)
	}

	type testCase struct {
		name    string
		id      ID
		filters []string
		correct bool
	}

	tests := []testCase{
		{
			name:    "Производим добавление фильтров пользователя",
			id:      1,
			filters: []string{"Name", "troll"},
			correct: true,
		},
		{
			name:    "Производим добавление фильтров не зарегистрированного пользователя",
			id:      2,
			filters: []string{"Name", "troll"},
			correct: false,
		},
	}

	for _, test := range tests {
		err := contextStore.AddFilters(test.id, test.filters)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestContextStorage_AddTags(t *testing.T) {
	contextStore := contextStorage.New()

	type StoreData struct {
		id ID
	}

	data := []StoreData{
		{id: 1},
	}

	for _, user := range data {
		contextStore.RegUser(user.id)
	}

	type testCase struct {
		name    string
		id      ID
		tags    []string
		correct bool
	}

	tests := []testCase{
		{
			name:    "Производим добавление тегов пользователя",
			id:      1,
			tags:    []string{"Work", "Family"},
			correct: true,
		},
		{
			name:    "Производим добавление тегов не зарегистрированного пользователя",
			id:      2,
			tags:    []string{"Work", "Family"},
			correct: false,
		},
	}

	for _, test := range tests {
		err := contextStore.AddTags(test.id, test.tags)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestContextStorage_ResetCtx(t *testing.T) {
	contextStore := contextStorage.New()

	type StoreData struct {
		id ID
	}

	data := []StoreData{
		{id: 1},
	}

	for _, user := range data {
		contextStore.RegUser(user.id)
	}

	type testCase struct {
		name    string
		id      ID
		correct bool
	}

	tests := []testCase{
		{
			name:    "Производим cброс контекста регнутого юзера",
			id:      1,
			correct: true,
		},
		{
			name:    "Производим cброс контекста не регнутого юзера",
			id:      2,
			correct: false,
		},
	}

	for _, test := range tests {
		err := contextStore.ResetCtx(test.id)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestContextStorage_UserContext(t *testing.T) {
	contextStore := contextStorage.New()

	type StoreData struct {
		id      ID
		URL     string
		Tags    []string
		Filters []string
	}

	data := []StoreData{
		{
			id:      1,
			URL:     "tbank.ru",
			Tags:    []string{},
			Filters: []string{},
		},
	}

	for _, user := range data {
		contextStore.RegUser(user.id)
		contextStore.AddUrl(user.id, user.URL)
		contextStore.AddTags(user.id, user.Tags)
		contextStore.AddFilters(user.id, user.Filters)
	}

	type testCase struct {
		name    string
		id      ID
		res     *ContextData
		correct bool
	}

	tests := []testCase{
		{
			name: "Получаем контекст регнутого юзера",
			id:   1,
			res: &ContextData{
				URL:     "tbank.ru",
				Filters: []string{},
				Tags:    []string{},
			},
			correct: true,
		},
		{
			name:    "Получаем контекст не регнутого юзера",
			id:      2,
			res:     nil,
			correct: false,
		},
	}

	for _, test := range tests {
		context, err := contextStore.UserContext(test.id)

		if test.correct {
			assert.NoError(t, err)
			assert.Equal(t, test.res, context)
		} else {
			assert.Error(t, err)
		}
	}
}
