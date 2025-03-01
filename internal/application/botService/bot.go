package botService

import (
	"linkTraccer/internal/domain/tgbot"
	"log"
	"time"
)

type Link = tgbot.Link
type Updates = tgbot.Updates
type ID = tgbot.ID

type TgClient interface {
	HandleUsersUpdates(offset, limit int) (Updates, error)
	SendMessage(userID int, text string) error
}

type CtxStorage interface {
	RegUser(id ID) error
	AddUrl(id ID, url string) error
	AddFilters(id ID, filters []string) error
	AddTags(id ID, tags []string) error
	ResetCtx(id ID) error
	UserContext(id ID) (*tgbot.ContextData, error)
}

type ScrapClient interface {
	RegUser(id ID) error
	AddLink(ID, *tgbot.ContextData) error
	RemoveLink(ID, Link) error
	UserLinks(ID) ([]Link, error)
}

type TgBot struct {
	offset      int
	limit       int
	client      TgClient
	ctxStore    CtxStorage
	scrapClient ScrapClient
}

const (
	Start   = "/start"   // Регистрация пользователя
	Help    = "/help"    // Вывод списка доступных команд.
	Track   = "/track"   // Начать отслеживание ссылки
	Untrack = "/untrack" //  Прекратить отслеживание ссылки.
	List    = "/list"    // Показать список отслеживаемых ссылок (cписок ссылок, полученных при /track)
)

// названия состояний бота

const (
	InitialState         = "init"    // В этом состоянии бот может принять только команду /start
	AnyRegisteredCommand = "comands" // В этом состоянии бот может выполнить любую команду
	RemoveLink           = "remove"  // В этом состоянии бот ждет ссылку для удаления, а так же может выполнить любую команду
	AddNewLink           = "link"    // В этом состоянии бот ждет ссылку, а так же может выполнить любую команду
	AddLinkTag           = "tag"     // В этом состоянии бот ждет тэг ссылки, а так же может выполнить любую команду
	AddLinkFilter        = "filter"  // В этом состоянии бот ждет фильтр ссылки, а так же может выполнить любую команду

)

// Сообщения пользователю

const (
	HelpMessage = `/start - поможет перезапустить бота
     /help - справка по всем командам 
     /track - добавить новую ссылку, на отслеживание
	 /untrack - удалить ссылку, за которой следите
	 /list - вернуть список всех отслеживаемых ссылок`

	FirstMessage = `Привет! Я бот, который может уведомлять тебя, об изменения в публичных репозиториях GitHub и о новых
ответах, на интересующий тебя вопрос StackOverflow` + "\n\n" + HelpMessage

	NoSavedLinks     = "У вас нет сохраненных ссылок😟"
	NotSaveThisLink  = "Вы не сохраняли такой ссылки❌"
	UnknownCommand   = "Я пока не знаю такой команды 😔. Введите /help"
	UntrackLink      = "Введите ссылку, которую хотите перестать отслеживать⬇️"
	TrackLink        = "Введите ссылку, которую хотите начать отслеживать⬇️"
	LinkDelete       = "Ссылка больше не отслеживаается✔️"
	AddLinkTagMsg    = "Добавьте тег для ссылки💬"
	AddLinkFilterMsg = "Введите фильтр для ссылки👁️‍🗨️"
	WrongLink        = "Ваша ссылка не поддерживается❌"
	GoodLink         = "Ссылка успешно сохранена✔️"
)

var StartTransition = tgbot.Transition{
	Event: Start,
	Dst:   AnyRegisteredCommand,
}

var HelpTransition = tgbot.Transition{
	Event: Help,
	Dst:   AnyRegisteredCommand,
}

var UntrackTransition = tgbot.Transition{
	Event: Untrack,
	Dst:   RemoveLink,
}

var RemoveTransition = tgbot.Transition{
	Event: tgbot.TextEvent,
	Dst:   AnyRegisteredCommand,
}

var ListTransition = tgbot.Transition{
	Event: List,
	Dst:   AnyRegisteredCommand,
}

var TrackTransition = tgbot.Transition{
	Event: Track,
	Dst:   AddNewLink,
}

var LinkTransition = tgbot.Transition{
	Event: tgbot.TextEvent,
	Dst:   AddLinkTag,
}

var TagTransition = tgbot.Transition{
	Event: tgbot.TextEvent,
	Dst:   AddLinkFilter,
}

var FilterTransition = tgbot.Transition{
	Event: tgbot.TextEvent,
	Dst:   AnyRegisteredCommand,
}

var commandTransition = tgbot.Transitions{
	StartTransition,
	HelpTransition,
	UntrackTransition,
	ListTransition,
	TrackTransition,
}

var states = tgbot.States{
	{
		Name: InitialState,
		Transitions: tgbot.Transitions{
			StartTransition,
		},
	},
	{
		Name:        AnyRegisteredCommand,
		Transitions: commandTransition,
	},
	{
		Name:        AddNewLink,
		Transitions: append(commandTransition, LinkTransition),
	},
	{
		Name:        AddLinkTag,
		Transitions: append(commandTransition, TagTransition),
	},
	{
		Name:        AddLinkFilter,
		Transitions: append(commandTransition, FilterTransition),
	},
	{
		Name:        RemoveLink,
		Transitions: append(commandTransition, RemoveTransition),
	},
}

type stateHandler func(TgClient, ScrapClient, CtxStorage, ID, tgbot.EventType) error

var mux = map[tgbot.StateType]stateHandler{
	InitialState:         RegHandler,
	AnyRegisteredCommand: CommandsHandler,
	RemoveLink:           LinkRemoveHandler,
	AddNewLink:           AddLinkHandler,
	AddLinkTag:           AddTagHandler,
	AddLinkFilter:        SaveLinkHandler,
}

func New(client TgClient, scrapClient ScrapClient, ctxStorage CtxStorage, limit int) *TgBot {
	return &TgBot{
		ctxStore:    ctxStorage,
		client:      client,
		limit:       limit,
		scrapClient: scrapClient,
	}
}

func (bot *TgBot) Start() {

	states, _ := tgbot.NewStateMachine(InitialState, states)

	for {

		updates, _ := bot.client.HandleUsersUpdates(bot.offset, bot.limit) // надо подумать что делать с ошибкой
		updatesNum := len(updates)

		if updatesNum > 0 {
			bot.changeOffset(updates[updatesNum-1].UpdateID + 1)

			log.Printf("Получено %d новых апдейтов", len(updates))

			for _, update := range updates {

				id := update.Msg.From.ID
				state := states.Current(id)
				err := mux[state](bot.client, bot.scrapClient, bot.ctxStore, id, update.Msg.Text)

				if err != nil {
					log.Println(err)
				}

				states.Transition(id, update.Msg.Text)
				//
				//bot.client.SendMessage(update.Msg.From.ID, update.Msg.Text)
			}
		}

		time.Sleep(time.Second * 5)
	}
}

func (bot *TgBot) changeOffset(newOffset int) {
	bot.offset = newOffset
}
