package botservice

import (
	"fmt"
	"linkTraccer/internal/domain/tgbot"
	"log/slog"
)

type TgClient interface {
	HandleUsersUpdates(offset, limit int) (tgbot.Updates, error)
	SendMessage(userID int64, text string) error
	SetBotCommands(data *tgbot.SetCommands) error
}

type CtxStorage interface {
	RegUser(id tgbot.ID) error
	AddURL(id tgbot.ID, url string) error
	AddFilters(id tgbot.ID, filters []string) error
	AddTags(id tgbot.ID, tags []string) error
	ResetCtx(id tgbot.ID) error
	UserContext(id tgbot.ID) (*tgbot.ContextData, error)
}

type ScrapClient interface {
	RegUser(id tgbot.ID) error
	AddLink(tgbot.ID, *tgbot.ContextData) error
	RemoveLink(tgbot.ID, tgbot.Link) error
	UserLinks(tgbot.ID) ([]tgbot.Link, error)
}

type TgBot struct {
	offset      int
	limit       int
	log         *slog.Logger
	states      *tgbot.StateMachine
	client      TgClient
	ctxStore    CtxStorage
	scrapClient ScrapClient
}

func New(client TgClient, scrapClient ScrapClient, ctxStorage CtxStorage, log *slog.Logger, limit int) *TgBot {
	return &TgBot{
		ctxStore:    ctxStorage,
		client:      client,
		limit:       limit,
		scrapClient: scrapClient,
		log:         log,
	}
}

func (bot *TgBot) Init() error {
	var err error

	bot.states, err = tgbot.NewStateMachine(InitialState, botStates())
	if err != nil {
		return err
	}

	if err := bot.setCommands(); err != nil {
		return err
	}

	return nil
}

func (bot *TgBot) CheckUsersMsg() {
	updates, _ := bot.client.HandleUsersUpdates(bot.offset, bot.limit) // надо подумать что делать с ошибкой
	updatesNum := len(updates)

	if updatesNum > 0 {
		bot.changeOffset(updates[updatesNum-1].UpdateID + 1)
		bot.log.Info(fmt.Sprintf("Получено %d новых апдейтов", len(updates)))

		for _, update := range updates {
			id := update.Msg.From.ID
			state := bot.states.Current(id)

			if _, ok := stateHandlers[state]; !ok {
				bot.log.Debug(fmt.Sprintf("у состояния %s, нет обработчика", state))

				continue
			}

			bot.log.Debug(fmt.Sprintf("у состояния %s, нет обработчика", state))

			err := stateHandlers[state](bot.client, bot.scrapClient, bot.ctxStore, id, update.Msg.Text)

			if err != nil {
				bot.log.Debug("ошибка при обработке состояния пользователя", "err", err.Error())
			}

			if _, err := bot.states.Transition(id, update.Msg.Text); err != nil {
				bot.log.Debug(fmt.Sprintf("ошибка при переходе из состояния %s", state))
			}
		}
	}
}

func (bot *TgBot) setCommands() error {
	commandsMsg := &tgbot.SetCommands{}

	for _, command := range commandsDescription {
		commandsMsg.Commands = append(commandsMsg.Commands,
			tgbot.BotCommand{Command: command[0], Description: command[1]})
	}

	if err := bot.client.SetBotCommands(commandsMsg); err != nil {
		return fmt.Errorf("ошибка при отправке запроса SetBotCommands: %w", err)
	}

	return nil
}

func (bot *TgBot) changeOffset(newOffset int) {
	bot.offset = newOffset
}
