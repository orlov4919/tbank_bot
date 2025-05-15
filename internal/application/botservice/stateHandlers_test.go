package botservice_test

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"linkTraccer/internal/application/botservice"
	"linkTraccer/internal/application/botservice/mocks"
	"linkTraccer/internal/domain/tgbot"
	"log/slog"
	"os"
	"testing"
)

const (
	testID         = 1
	unknownCommand = "/helloword"
	botLimit       = 1
	link           = "http://tbank.com"
	linkTag        = "Hello Word"
	linkFilters    = "link filter"
)

var (
	errTest  = errors.New("ошибка для тест")
	links    = []tgbot.Link{"tbank.ru"}
	logLevel = slog.LevelDebug

	logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
)

func TestTgBot_Commands(t *testing.T) {
	tgWithErr := mocks.NewTgClient(t)
	tgWithoutErr := mocks.NewTgClient(t)

	scrapWithError := mocks.NewScrapClient(t)
	scrapWithoutLinks := mocks.NewScrapClient(t)

	store := mocks.NewCtxStorage(t)

	notEmtyCache := mocks.NewCacheStorage(t)
	emptyCache := mocks.NewCacheStorage(t)

	tgWithoutErr.On("SendMessage", mock.Anything, mock.Anything).Return(nil)
	tgWithErr.On("SendMessage", mock.Anything, mock.Anything).Return(errTest)

	scrapWithError.On("UserLinks", mock.Anything).Return(nil, errTest)
	scrapWithoutLinks.On("UserLinks", mock.Anything).Return([]tgbot.Link{}, nil)

	emptyCache.On("GetUserLinks", mock.Anything).Return("", errTest)
	notEmtyCache.On("GetUserLinks", mock.Anything).Return("hello word", nil)

	type testCase struct {
		name    string
		tg      botservice.TgClient
		cache   botservice.CacheStorage
		scrap   botservice.ScrapClient
		event   tgbot.Event
		correct bool
	}

	tests := []testCase{
		{
			name:    "Ошибка при отправке приветственного сообщения",
			tg:      tgWithErr,
			cache:   nil,
			scrap:   scrapWithoutLinks,
			event:   botservice.Start,
			correct: false,
		},
		{
			name:    "произошла ошибка при отправке сообщения с командами",
			tg:      tgWithErr,
			scrap:   scrapWithoutLinks,
			event:   botservice.Help,
			correct: false,
		},
		{
			name:    "кеш пользователя пустой, пытаемся получить данный из БД, но происходит ошибка",
			tg:      tgWithErr,
			cache:   emptyCache,
			scrap:   scrapWithError,
			event:   botservice.List,
			correct: false,
		},
		{
			name:    "при отправке о том что у пользователя нет ссылок, происходит ошибка",
			tg:      tgWithErr,
			cache:   emptyCache,
			scrap:   scrapWithoutLinks,
			event:   botservice.List,
			correct: false,
		},
		{
			name:    "получение ссылок из кеша, успешная отправка сообщения в тг",
			tg:      tgWithoutErr,
			cache:   notEmtyCache,
			scrap:   scrapWithoutLinks,
			event:   botservice.List,
			correct: true,
		},
		{
			name:    "ошибка при отправке сообщения для удаления ссылки",
			tg:      tgWithErr,
			scrap:   scrapWithError,
			event:   botservice.Untrack,
			correct: false,
		},
		{
			name:    "ошибка при отправке сообщения с добавлением ссылки",
			tg:      tgWithErr,
			scrap:   scrapWithError,
			event:   botservice.Track,
			correct: false,
		},
		{
			name:    "в боте нет обработчика для такой команды",
			tg:      tgWithErr,
			scrap:   scrapWithError,
			event:   unknownCommand,
			correct: false,
		},
	}

	for _, test := range tests {
		bot := botservice.New(test.tg, test.scrap, store, test.cache, logger, botLimit)
		err := bot.Commands(testID, test.event)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestTgBot_RegHandler(t *testing.T) {
	tgWithErr := mocks.NewTgClient(t)
	tgWithoutErr := mocks.NewTgClient(t)

	cache := mocks.NewCacheStorage(t)

	storeWithErr := mocks.NewCtxStorage(t)
	storeWithoutErr := mocks.NewCtxStorage(t)

	scrapWithErr := mocks.NewScrapClient(t)
	scrapWithoutErr := mocks.NewScrapClient(t)

	storeWithErr.On("RegUser", mock.Anything).Return(errTest)
	storeWithoutErr.On("RegUser", mock.Anything).Return(nil)

	scrapWithErr.On("RegUser", mock.Anything).Return(errTest)
	scrapWithoutErr.On("RegUser", mock.Anything).Return(nil)

	tgWithErr.On("SendMessage", mock.Anything, mock.Anything).Return(errTest)
	tgWithoutErr.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

	type testCase struct {
		name     string
		tg       botservice.TgClient
		scrap    botservice.ScrapClient
		ctxStore botservice.CtxStorage
		event    tgbot.Event
		correct  bool
	}

	tests := []testCase{
		{
			name:     "ошибка при добавлении id в контекстное хранилище",
			tg:       tgWithErr,
			scrap:    scrapWithErr,
			ctxStore: storeWithErr,
			event:    botservice.Start,
			correct:  false,
		},
		{
			name:     "ошибка при отправке приветственного сообщения",
			tg:       tgWithErr,
			scrap:    scrapWithErr,
			ctxStore: storeWithoutErr,
			event:    botservice.Start,
			correct:  false,
		},
		{
			name:     "ошибка при регистрации пользователя в скрапере",
			tg:       tgWithoutErr,
			scrap:    scrapWithErr,
			ctxStore: storeWithoutErr,
			event:    botservice.Start,
			correct:  false,
		},
		{
			name:     "успешная регистрация пользователя",
			tg:       tgWithoutErr,
			scrap:    scrapWithoutErr,
			ctxStore: storeWithoutErr,
			event:    botservice.Start,
			correct:  true,
		},
	}

	for _, test := range tests {
		tgBot := botservice.New(test.tg, test.scrap, test.ctxStore, cache, logger, botLimit)
		err := tgBot.RegHandler(testID, test.event)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestTgBot_CommandsHandler(t *testing.T) {
	tgWithErr := mocks.NewTgClient(t)
	tgWithoutErr := mocks.NewTgClient(t)

	cache := mocks.NewCacheStorage(t)

	scrap := mocks.NewScrapClient(t)

	store := mocks.NewCtxStorage(t)

	tgWithErr.On("SendMessage", mock.Anything, mock.Anything).Return(errTest)
	tgWithoutErr.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

	type testCase struct {
		name     string
		tg       botservice.TgClient
		scrap    botservice.ScrapClient
		ctxStore botservice.CtxStorage
		event    tgbot.Event
		correct  bool
	}

	tests := []testCase{
		{
			name:     "передаем не обрабатываемую команду и падаем при отправке сообщения",
			tg:       tgWithErr,
			scrap:    scrap,
			ctxStore: store,
			event:    unknownCommand,
			correct:  false,
		},
		{
			name:     "передаем команду на запуск, без ошибок",
			tg:       tgWithoutErr,
			scrap:    scrap,
			ctxStore: store,
			event:    botservice.Start,
			correct:  true,
		},
	}

	for _, test := range tests {
		tgBot := botservice.New(test.tg, test.scrap, test.ctxStore, cache, logger, botLimit)
		err := tgBot.CommandsHandler(testID, test.event)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestTgBot_LinkRemoveHandler(t *testing.T) {
	scrapWithoutLinks := mocks.NewScrapClient(t)
	scrapWithErr := mocks.NewScrapClient(t)
	scrapWithoutErr := mocks.NewScrapClient(t)

	store := mocks.NewCtxStorage(t)

	tgWithErr := mocks.NewTgClient(t)
	tgWithoutErr := mocks.NewTgClient(t)

	cacheWithErr := mocks.NewCacheStorage(t)

	cacheWithErr.On("InvalidateUserCache", mock.Anything).Return(errTest)

	scrapWithoutErr.On("RemoveLink", mock.Anything, mock.Anything).Return(nil)
	scrapWithoutLinks.On("RemoveLink", mock.Anything, mock.Anything).Return(tgbot.LinkNotExist)
	scrapWithErr.On("RemoveLink", mock.Anything, mock.Anything).Return(errTest)

	tgWithErr.On("SendMessage", mock.Anything, mock.Anything).Return(errTest)
	tgWithoutErr.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

	type testCase struct {
		name    string
		tg      botservice.TgClient
		scrap   botservice.ScrapClient
		cache   botservice.CacheStorage
		event   tgbot.Event
		correct bool
	}

	tests := []testCase{
		{
			name:    "пришла команда на добавление ссылки",
			tg:      tgWithErr,
			cache:   cacheWithErr,
			scrap:   scrapWithoutLinks,
			event:   botservice.Track,
			correct: false,
		},
		{
			name:    "ссылка, которую пытаемся удалить не была сохранена, отправляем об этом сообщение",
			tg:      tgWithoutErr,
			cache:   cacheWithErr,
			scrap:   scrapWithoutLinks,
			event:   link,
			correct: true,
		},
		{
			name:    "ошибка при удалении ссылки из скрапера",
			tg:      tgWithErr,
			cache:   cacheWithErr,
			scrap:   scrapWithErr,
			event:   link,
			correct: false,
		},
		{
			name:    "ошибка при удалении ссылки из кеша, успешная отправка сообщения об удалении",
			tg:      tgWithoutErr,
			cache:   cacheWithErr,
			scrap:   scrapWithoutErr,
			event:   link,
			correct: true,
		},
	}

	for _, test := range tests {
		tgBot := botservice.New(test.tg, test.scrap, store, test.cache, logger, botLimit)
		err := tgBot.LinkRemoveHandler(testID, test.event)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestTgBot_AddLinkHandler(t *testing.T) {
	scrap := mocks.NewScrapClient(t)

	cache := mocks.NewCacheStorage(t)

	storeWithErr := mocks.NewCtxStorage(t)
	storeWithoutErr := mocks.NewCtxStorage(t)

	tgWithoutErr := mocks.NewTgClient(t)

	storeWithErr.On("AddURL", mock.Anything, mock.Anything).Return(errTest)
	storeWithoutErr.On("AddURL", mock.Anything, mock.Anything).Return(nil)

	tgWithoutErr.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

	type testCase struct {
		name     string
		tg       botservice.TgClient
		scrap    botservice.ScrapClient
		ctxStore botservice.CtxStorage
		event    tgbot.Event
		correct  bool
	}

	tests := []testCase{
		{
			name:     "пришла команда на добавление ссылки",
			tg:       tgWithoutErr,
			scrap:    scrap,
			ctxStore: storeWithErr,
			event:    botservice.Track,
			correct:  true,
		},
		{
			name:     "ошибка при добавлении ссылки в контекстное хранилище",
			tg:       tgWithoutErr,
			scrap:    scrap,
			ctxStore: storeWithErr,
			event:    link,
			correct:  false,
		},
		{
			name:     "добавление ссылки прошло успешно",
			tg:       tgWithoutErr,
			scrap:    scrap,
			ctxStore: storeWithoutErr,
			event:    link,
			correct:  true,
		},
	}

	for _, test := range tests {
		tgBot := botservice.New(test.tg, test.scrap, test.ctxStore, cache, logger, botLimit)
		err := tgBot.AddLinkHandler(testID, test.event)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestTgBot_AddTagHandler(t *testing.T) {
	scrap := mocks.NewScrapClient(t)

	cache := mocks.NewCacheStorage(t)

	storeWithErr := mocks.NewCtxStorage(t)
	storeWithoutErr := mocks.NewCtxStorage(t)

	tgWithoutErr := mocks.NewTgClient(t)

	storeWithErr.On("AddTags", mock.Anything, mock.Anything).Return(errTest)
	storeWithoutErr.On("AddTags", mock.Anything, mock.Anything).Return(nil)

	tgWithoutErr.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

	type testCase struct {
		name     string
		tg       botservice.TgClient
		scrap    botservice.ScrapClient
		ctxStore botservice.CtxStorage
		event    tgbot.Event
		correct  bool
	}

	tests := []testCase{
		{
			name:     "пришла команда на добавление ссылки",
			tg:       tgWithoutErr,
			scrap:    scrap,
			ctxStore: storeWithErr,
			event:    botservice.Track,
			correct:  true,
		},
		{
			name:     "ошибка при добавлении тегов в контекстное хранилище",
			tg:       tgWithoutErr,
			scrap:    scrap,
			ctxStore: storeWithErr,
			event:    linkTag,
			correct:  false,
		},
		{
			name:     "добавление тегов прошло успешно",
			tg:       tgWithoutErr,
			scrap:    scrap,
			ctxStore: storeWithoutErr,
			event:    linkTag,
			correct:  true,
		},
	}

	for _, test := range tests {
		tgBot := botservice.New(test.tg, test.scrap, test.ctxStore, cache, logger, botLimit)
		err := tgBot.AddTagHandler(testID, test.event)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestTgBot_SaveLinkHandler(t *testing.T) {
	tgWithoutErr := mocks.NewTgClient(t)

	cacheWithErr := mocks.NewCacheStorage(t)

	storeWithAddErr := mocks.NewCtxStorage(t)
	storeWithContextErr := mocks.NewCtxStorage(t)
	storeWithoutErr := mocks.NewCtxStorage(t)

	scrapWithLinkNotSupport := mocks.NewScrapClient(t)
	scrapWithInternalErr := mocks.NewScrapClient(t)
	scrapWithoutErr := mocks.NewScrapClient(t)

	tgWithoutErr.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

	storeWithAddErr.On("AddFilters", mock.Anything, mock.Anything).Return(errTest)
	storeWithContextErr.On("AddFilters", mock.Anything, mock.Anything).Return(nil)
	storeWithContextErr.On("UserContext", mock.Anything, mock.Anything).
		Return(nil, errTest)

	storeWithoutErr.On("AddFilters", mock.Anything, mock.Anything).Return(nil)
	storeWithoutErr.On("UserContext", mock.Anything, mock.Anything).
		Return(&tgbot.ContextData{}, nil)

	scrapWithLinkNotSupport.On("AddLink", mock.Anything, mock.Anything).Return(tgbot.LinkNotSupport)
	scrapWithInternalErr.On("AddLink", mock.Anything, mock.Anything).Return(errTest)
	scrapWithoutErr.On("AddLink", mock.Anything, mock.Anything).Return(nil)

	cacheWithErr.On("InvalidateUserCache", mock.Anything).Return(errTest)

	type testCase struct {
		name     string
		tg       botservice.TgClient
		scrap    botservice.ScrapClient
		ctxStore botservice.CtxStorage
		event    tgbot.Event
		correct  bool
	}

	tests := []testCase{
		{
			name:     "пришла команда на добавление новой ссылки",
			tg:       tgWithoutErr,
			scrap:    scrapWithLinkNotSupport,
			ctxStore: storeWithAddErr,
			event:    botservice.Track,
			correct:  true,
		},
		{
			name:     "Произошла ошибка при добавлении фильтров в контекстное хранилище",
			tg:       tgWithoutErr,
			scrap:    scrapWithLinkNotSupport,
			ctxStore: storeWithAddErr,
			event:    linkFilters,
			correct:  false,
		},
		{
			name:     "Произошла ошибка при получении контекста",
			tg:       tgWithoutErr,
			scrap:    scrapWithLinkNotSupport,
			ctxStore: storeWithContextErr,
			event:    linkFilters,
			correct:  false,
		},
		{
			name:     "Добавляемая ссылка не поддерживается",
			tg:       tgWithoutErr,
			scrap:    scrapWithLinkNotSupport,
			ctxStore: storeWithoutErr,
			event:    linkFilters,
			correct:  true,
		},
		{
			name:     "Ошибка в скрапере при добавлении ссылки",
			tg:       tgWithoutErr,
			scrap:    scrapWithInternalErr,
			ctxStore: storeWithoutErr,
			event:    linkFilters,
			correct:  false,
		},
		{
			name:     "Добавление ссылки произошло успешно",
			tg:       tgWithoutErr,
			scrap:    scrapWithoutErr,
			ctxStore: storeWithoutErr,
			event:    linkFilters,
			correct:  true,
		},
	}

	for _, test := range tests {
		tgBot := botservice.New(test.tg, test.scrap, test.ctxStore, cacheWithErr, logger, botLimit)
		err := tgBot.SaveLinkHandler(testID, test.event)

		if test.correct {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}
