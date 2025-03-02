package botservice_test

import (
	"errors"
	"linkTraccer/internal/application/botservice"
	"linkTraccer/internal/application/botservice/mocks"
	"linkTraccer/internal/domain/tgbot"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	ID           = 1
	wrongCommand = "/sell"
)

var (
	errTest = errors.New("ошибка для тест")
	links   = []Link{"tbank.ru"}
)

type Link = tgbot.Link
type TgClient = botservice.TgClient
type ScrapClient = botservice.ScrapClient
type CtxStorage = botservice.CtxStorage
type userID = tgbot.ID
type EventType = tgbot.EventType

func TestCommandsHandler(t *testing.T) {
	wrongClient := mocks.NewTgClient(t)
	goodClient := mocks.NewTgClient(t)
	store := mocks.NewCtxStorage(t)
	scrapWithError := mocks.NewScrapClient(t)
	scrapWithoutLinks := mocks.NewScrapClient(t)
	scrapWithLinks := mocks.NewScrapClient(t)

	goodClient.On("SendMessage", mock.Anything, mock.Anything).Return(nil)
	wrongClient.On("SendMessage", mock.Anything, mock.Anything).Return(errTest)
	scrapWithError.On("UserLinks", mock.Anything).Return(nil, errTest)
	scrapWithoutLinks.On("UserLinks", mock.Anything).Return([]Link{}, nil)
	scrapWithLinks.On("UserLinks", mock.Anything).Return(links, nil)

	type testCase struct {
		name       string
		client     TgClient
		scrap      ScrapClient
		ctxStorage CtxStorage
		id         userID
		event      EventType
		correct    bool
	}

	tests := []testCase{
		{
			name:       "Произошла ошибка при отправке сообщений пользователю в состоянии " + botservice.Start,
			client:     wrongClient,
			scrap:      scrapWithoutLinks,
			ctxStorage: store,
			id:         ID,
			event:      botservice.Start,
			correct:    false,
		},
		{
			name:       "Произошла ошибка при отправке сообщений пользователю, команда: " + botservice.Help,
			client:     wrongClient,
			scrap:      scrapWithoutLinks,
			ctxStorage: store,
			id:         ID,
			event:      botservice.Help,
			correct:    false,
		},
		{
			name:       "скрапер с ошибкой, команда:  " + botservice.List,
			client:     wrongClient,
			scrap:      scrapWithError,
			ctxStorage: store,
			id:         ID,
			event:      botservice.List,
			correct:    false,
		},
		{
			name:       "скрапер без ссылок, команда:  " + botservice.List,
			client:     wrongClient,
			scrap:      scrapWithoutLinks,
			ctxStorage: store,
			id:         ID,
			event:      botservice.List,
			correct:    false,
		},
		{
			name:       "скрапер c ссылкой, команда:  " + botservice.List,
			client:     wrongClient,
			scrap:      scrapWithLinks,
			ctxStorage: store,
			id:         ID,
			event:      botservice.List,
			correct:    false,
		},
		{
			name:       "скрапер с ошибкой, команда:  " + botservice.List,
			client:     wrongClient,
			scrap:      scrapWithError,
			ctxStorage: store,
			id:         ID,
			event:      botservice.List,
			correct:    false,
		},
		{
			name:       "ошибка при отправке сообщения пользователю, команда:  " + botservice.Untrack,
			client:     wrongClient,
			scrap:      scrapWithError,
			ctxStorage: store,
			id:         ID,
			event:      botservice.Untrack,
			correct:    false,
		},
		{
			name:       "ошибка при отправке сообщения пользователю, команда:  " + botservice.Track,
			client:     wrongClient,
			scrap:      scrapWithError,
			ctxStorage: store,
			id:         ID,
			event:      botservice.Track,
			correct:    false,
		},
		{
			name:       "случай без ошибок",
			client:     goodClient,
			scrap:      scrapWithError,
			ctxStorage: store,
			id:         ID,
			event:      botservice.Start,
			correct:    true,
		},
	}

	for _, test := range tests {
		err := botservice.CommandsHandler(test.client, test.scrap, test.ctxStorage, test.id, test.event)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestSaveLinkHandler(t *testing.T) {
	wrongClient := mocks.NewTgClient(t)
	goodClient := mocks.NewTgClient(t)
	storeWithAddErr := mocks.NewCtxStorage(t)
	storeWithContextErr := mocks.NewCtxStorage(t)
	storeWithoutErr := mocks.NewCtxStorage(t)
	scrapWithErr := mocks.NewScrapClient(t)
	scrapWithoutErr := mocks.NewScrapClient(t)

	goodClient.On("SendMessage", mock.Anything, mock.Anything).Return(nil)
	wrongClient.On("SendMessage", mock.Anything, mock.Anything).Return(errTest)
	storeWithAddErr.On("AddFilters", mock.Anything, mock.Anything).Return(errTest)
	storeWithContextErr.On("AddFilters", mock.Anything, mock.Anything).Return(nil)
	storeWithContextErr.On("UserContext", mock.Anything, mock.Anything).
		Return(nil, errTest)

	storeWithoutErr.On("AddFilters", mock.Anything, mock.Anything).Return(nil)
	storeWithoutErr.On("UserContext", mock.Anything, mock.Anything).
		Return(&tgbot.ContextData{}, nil)

	scrapWithErr.On("AddLink", mock.Anything, mock.Anything).Return(errTest)
	scrapWithoutErr.On("AddLink", mock.Anything, mock.Anything).Return(nil)

	type testCase struct {
		name       string
		client     TgClient
		scrap      ScrapClient
		ctxStorage CtxStorage
		id         userID
		event      EventType
		correct    bool
	}

	tests := []testCase{
		{
			name:       "Произошла ошибка при добавлении фильтров",
			client:     wrongClient,
			scrap:      scrapWithErr,
			ctxStorage: storeWithAddErr,
			id:         ID,
			event:      botservice.Start,
			correct:    false,
		},
		{
			name:       "Произошла ошибка при получении контекста",
			client:     wrongClient,
			scrap:      scrapWithErr,
			ctxStorage: storeWithContextErr,
			id:         ID,
			event:      botservice.Start,
			correct:    false,
		},
		{
			name:       "Произошла ошибка при получении контекста",
			client:     wrongClient,
			scrap:      scrapWithErr,
			ctxStorage: storeWithContextErr,
			id:         ID,
			event:      botservice.Start,
			correct:    false,
		},
		{
			name:       "Произошла ошибка при добавлении ссылки",
			client:     wrongClient,
			scrap:      scrapWithErr,
			ctxStorage: storeWithoutErr,
			id:         ID,
			event:      botservice.Start,
			correct:    false,
		},
		{
			name:       "Произошла ошибка при отправке сообщения",
			client:     wrongClient,
			scrap:      scrapWithoutErr,
			ctxStorage: storeWithoutErr,
			id:         ID,
			event:      botservice.Start,
			correct:    false,
		},
		{
			name:       "Тест без ошибок",
			client:     goodClient,
			scrap:      scrapWithoutErr,
			ctxStorage: storeWithoutErr,
			id:         ID,
			event:      botservice.Start,
			correct:    true,
		},
	}

	for _, test := range tests {
		err := botservice.SaveLinkHandler(test.client, test.scrap, test.ctxStorage, test.id, test.event)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestAddTagHandler(t *testing.T) {
	scrap := mocks.NewScrapClient(t)
	storeWithErr := mocks.NewCtxStorage(t)
	storeWithoutErr := mocks.NewCtxStorage(t)
	badClient := mocks.NewTgClient(t)
	goodClient := mocks.NewTgClient(t)

	storeWithErr.On("AddTags", mock.Anything, mock.Anything).Return(errTest)
	storeWithoutErr.On("AddTags", mock.Anything, mock.Anything).Return(nil)
	badClient.On("SendMessage", mock.Anything, mock.Anything).Return(errTest)
	goodClient.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

	type testCase struct {
		name       string
		client     TgClient
		scrap      ScrapClient
		ctxStorage CtxStorage
		id         userID
		event      EventType
		correct    bool
	}

	tests := []testCase{
		{
			name:       "ошибка при добавлении тегов",
			client:     badClient,
			scrap:      scrap,
			ctxStorage: storeWithErr,
			id:         ID,
			event:      botservice.Start,
			correct:    false,
		},
		{
			name:       "ошибка при отправке сообщения",
			client:     badClient,
			scrap:      scrap,
			ctxStorage: storeWithoutErr,
			id:         ID,
			event:      botservice.Start,
			correct:    false,
		},
		{
			name:       "без ошибок",
			client:     goodClient,
			scrap:      scrap,
			ctxStorage: storeWithoutErr,
			id:         ID,
			event:      botservice.Start,
			correct:    true,
		},
	}

	for _, test := range tests {
		err := botservice.AddTagHandler(test.client, test.scrap, test.ctxStorage, test.id, test.event)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestAddLinkHandler(t *testing.T) {
	scrap := mocks.NewScrapClient(t)
	storeWithErr := mocks.NewCtxStorage(t)
	storeWithoutErr := mocks.NewCtxStorage(t)
	badClient := mocks.NewTgClient(t)
	goodClient := mocks.NewTgClient(t)

	storeWithErr.On("AddURL", mock.Anything, mock.Anything).Return(errTest)
	storeWithoutErr.On("AddURL", mock.Anything, mock.Anything).Return(nil)

	badClient.On("SendMessage", mock.Anything, mock.Anything).Return(errTest)
	goodClient.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

	type testCase struct {
		name       string
		client     TgClient
		scrap      ScrapClient
		ctxStorage CtxStorage
		id         userID
		event      EventType
		correct    bool
	}

	tests := []testCase{
		{
			name:       "ошибка при добавлении урла",
			client:     badClient,
			scrap:      scrap,
			ctxStorage: storeWithErr,
			id:         ID,
			event:      botservice.Start,
			correct:    false,
		},
		{
			name:       "ошибка при отправке сообщения",
			client:     badClient,
			scrap:      scrap,
			ctxStorage: storeWithoutErr,
			id:         ID,
			event:      botservice.Start,
			correct:    false,
		},
		{
			name:       "без ошибок",
			client:     goodClient,
			scrap:      scrap,
			ctxStorage: storeWithoutErr,
			id:         ID,
			event:      botservice.Start,
			correct:    true,
		},
	}

	for _, test := range tests {
		err := botservice.AddLinkHandler(test.client, test.scrap, test.ctxStorage, test.id, test.event)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestLinkRemoveHandler(t *testing.T) {
	scrapWithErr := mocks.NewScrapClient(t)
	scrapWithoutErr := mocks.NewScrapClient(t)

	storeWithErr := mocks.NewCtxStorage(t)
	storeWithoutErr := mocks.NewCtxStorage(t)

	badClient := mocks.NewTgClient(t)
	goodClient := mocks.NewTgClient(t)

	scrapWithErr.On("RemoveLink", mock.Anything, mock.Anything).Return(errTest)
	scrapWithoutErr.On("RemoveLink", mock.Anything, mock.Anything).Return(nil)

	storeWithErr.On("AddURL", mock.Anything, mock.Anything).Return(errTest)
	storeWithoutErr.On("AddURL", mock.Anything, mock.Anything).Return(nil)

	badClient.On("SendMessage", mock.Anything, mock.Anything).Return(errTest)
	goodClient.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

	type testCase struct {
		name       string
		client     TgClient
		scrap      ScrapClient
		ctxStorage CtxStorage
		id         userID
		event      EventType
		correct    bool
	}

	tests := []testCase{
		{
			name:       "ошибка при добавлении урла",
			client:     badClient,
			scrap:      scrapWithErr,
			ctxStorage: storeWithErr,
			id:         ID,
			event:      botservice.Start,
			correct:    false,
		},
		{
			name:       "ошибка при удалении урла",
			client:     badClient,
			scrap:      scrapWithErr,
			ctxStorage: storeWithoutErr,
			id:         ID,
			event:      botservice.Start,
			correct:    false,
		},
		{
			name:       "ошибка при отправке сообщения",
			client:     badClient,
			scrap:      scrapWithoutErr,
			ctxStorage: storeWithoutErr,
			id:         ID,
			event:      botservice.Start,
			correct:    false,
		},
		{
			name:       "без ошибок",
			client:     goodClient,
			scrap:      scrapWithoutErr,
			ctxStorage: storeWithoutErr,
			id:         ID,
			event:      botservice.Start,
			correct:    true,
		},
	}

	for _, test := range tests {
		err := botservice.LinkRemoveHandler(test.client, test.scrap, test.ctxStorage, test.id, test.event)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestCommandsStateHandler(t *testing.T) {
	badClient := mocks.NewTgClient(t)
	goodClient := mocks.NewTgClient(t)
	scrap := mocks.NewScrapClient(t)
	store := mocks.NewCtxStorage(t)

	badClient.On("SendMessage", mock.Anything, mock.Anything).Return(errTest)
	goodClient.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

	type testCase struct {
		name       string
		client     TgClient
		scrap      ScrapClient
		ctxStorage CtxStorage
		id         userID
		event      EventType
		correct    bool
	}

	tests := []testCase{
		{
			name:       "ошибка при отправке сообщений",
			client:     badClient,
			scrap:      scrap,
			ctxStorage: store,
			id:         ID,
			event:      botservice.Start,
			correct:    false,
		},
		{
			name:       "без ошибок",
			client:     goodClient,
			scrap:      scrap,
			ctxStorage: store,
			id:         ID,
			event:      botservice.Start,
			correct:    true,
		},
	}

	for _, test := range tests {
		err := botservice.CommandsStateHandler(test.client, test.scrap, test.ctxStorage, test.id, test.event)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestRegHandler(t *testing.T) {
	clientWithErr := mocks.NewTgClient(t)
	clientWithoutErr := mocks.NewTgClient(t)
	storeWithErr := mocks.NewCtxStorage(t)
	scrapWithErr := mocks.NewScrapClient(t)
	scrapWithoutErr := mocks.NewScrapClient(t)
	storeWithoutErr := mocks.NewCtxStorage(t)

	storeWithErr.On("RegUser", mock.Anything).Return(errTest)
	storeWithoutErr.On("RegUser", mock.Anything).Return(nil)
	scrapWithErr.On("RegUser", mock.Anything).Return(errTest)
	scrapWithoutErr.On("RegUser", mock.Anything).Return(nil)
	clientWithErr.On("SendMessage", mock.Anything, mock.Anything).Return(errTest)
	clientWithoutErr.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

	type testCase struct {
		name       string
		client     TgClient
		scrap      ScrapClient
		ctxStorage CtxStorage
		id         userID
		event      EventType
		correct    bool
	}

	tests := []testCase{
		{
			name:       "ошибка при добавлении id в контекст",
			client:     clientWithErr,
			scrap:      scrapWithErr,
			ctxStorage: storeWithErr,
			id:         ID,
			event:      botservice.Start,
			correct:    false,
		},
		{
			name:       "ошибка при регистрации пользовтеля в скрапере",
			client:     clientWithErr,
			scrap:      scrapWithErr,
			ctxStorage: storeWithoutErr,
			id:         ID,
			event:      botservice.Start,
			correct:    false,
		},
		{
			name:       "ошибка при отправлении сообщения",
			client:     clientWithErr,
			scrap:      scrapWithoutErr,
			ctxStorage: storeWithoutErr,
			id:         ID,
			event:      botservice.Start,
			correct:    false,
		},
		{
			name:       "без ошибок",
			client:     clientWithoutErr,
			scrap:      scrapWithoutErr,
			ctxStorage: storeWithoutErr,
			id:         ID,
			event:      botservice.Start,
			correct:    true,
		},
	}

	for _, test := range tests {
		err := botservice.RegHandler(test.client, test.scrap, test.ctxStorage, test.id, test.event)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}
