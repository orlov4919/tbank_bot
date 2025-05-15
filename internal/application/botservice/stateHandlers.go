package botservice

import (
	"errors"
	"fmt"
	"linkTraccer/internal/domain/tgbot"
	"strings"
)

func (bot *TgBot) RegHandler(id tgbot.ID, _ tgbot.Event) error {
	if err := bot.ctxStore.RegUser(id); err != nil {
		return fmt.Errorf("при регистрации в хранилище контекстной информации возникла ошибка: %w", err)
	}

	if err := bot.tg.SendMessage(id, FirstMessage); err != nil {
		return fmt.Errorf("при отправке %s произошла ошибка: %w", FirstMessage, err)
	}

	if err := bot.scrap.RegUser(id); err != nil {
		return fmt.Errorf("при регистрации пользователя произошла ошибка: %w", err)
	}

	return nil
}

func (bot *TgBot) CommandsHandler(id tgbot.ID, event tgbot.Event) error {
	err := bot.Commands(id, event)
	if err != nil {
		return bot.sendMessage(id, UnknownCommand)
	}

	return nil
}

func (bot *TgBot) LinkRemoveHandler(id tgbot.ID, event tgbot.Event) error {
	err := bot.Commands(id, event)
	if err == nil || !errors.Is(err, ErrCommandNotFound) {
		return err
	}

	err = bot.scrap.RemoveLink(id, event)
	if errors.Is(err, tgbot.LinkNotExist) {
		return bot.sendMessage(id, NotSaveThisLink)
	}

	if err != nil {
		return err
	}

	if err := bot.cache.InvalidateUserCache(id); err != nil {
		bot.log.Error("ошибка инвалидации кеша, при удалении ссылки", "err", err.Error())
	}

	return bot.sendMessage(id, LinkDeleted)
}

func (bot *TgBot) AddLinkHandler(id tgbot.ID, event tgbot.Event) error {
	err := bot.Commands(id, event)
	if err == nil || !errors.Is(err, ErrCommandNotFound) {
		return err
	}

	if err := bot.ctxStore.AddURL(id, event); err != nil {
		return fmt.Errorf("при добавлении ссылки в контекстное хранилище, произошла ошибка :%w", err)
	}

	return bot.sendMessage(id, AddLinkTagMsg)
}

func (bot *TgBot) AddTagHandler(id tgbot.ID, event tgbot.Event) error {
	err := bot.Commands(id, event)
	if err == nil || !errors.Is(err, ErrCommandNotFound) {
		return err
	}

	if err := bot.ctxStore.AddTags(id, []string{event}); err != nil {
		return fmt.Errorf("при добавлении тегов в контекстное хранилище, произошла ошибка :%w", err)
	}

	return bot.sendMessage(id, AddLinkFilterMsg)
}

func (bot *TgBot) SaveLinkHandler(id tgbot.ID, event tgbot.Event) error {
	err := bot.Commands(id, event)
	if err == nil || !errors.Is(err, ErrCommandNotFound) {
		return nil
	}

	if err := bot.ctxStore.AddFilters(id, []string{event}); err != nil {
		return fmt.Errorf("при добавлении фильтров в контекстное хранилище произошла ошибка: %w", err)
	}

	userContext, err := bot.ctxStore.UserContext(id)
	if err != nil {
		return fmt.Errorf("ошибка при сохраненни, при получении контекстной информации произошла ошибка: %w", err)
	}

	err = bot.scrap.AddLink(id, userContext)
	if errors.Is(err, tgbot.LinkNotSupport) {
		return bot.sendMessage(id, WrongLink)
	}

	if err != nil {
		return err
	}

	if err := bot.cache.InvalidateUserCache(id); err != nil {
		bot.log.Error("ошибка инвалидации кеша, при добавлении ссылки", "err", err.Error())
	}

	return bot.sendMessage(id, GoodLink)
}

func (bot *TgBot) Commands(id tgbot.ID, command tgbot.Event) error {
	switch command {
	case Start:
		return bot.sendMessage(id, FirstMessage)
	case Help:
		return bot.sendMessage(id, HelpMessage)
	case List:
		links, err := bot.cache.GetUserLinks(id)
		if err != nil {
			bot.log.Error("не удалось получить список ссылок из кеша", "err", err.Error())

			userLinks, err := bot.scrap.UserLinks(id)
			if err != nil {
				return err
			}

			links = formatLinksMsg(userLinks)

			if err := bot.cache.SetUserLinks(id, links); err != nil {
				bot.log.Error("ошибка при кешировании ссылок пользователя", "err", err.Error())
			}
		}

		if len(links) == 0 {
			return bot.sendMessage(id, NoSavedLinks)
		}

		return bot.sendMessage(id, links)

	case Untrack:
		return bot.sendMessage(id, UntrackLink)
	case Track:
		return bot.sendMessage(id, TrackLink)
	default:
		return ErrCommandNotFound
	}
}

func (bot *TgBot) sendMessage(id tgbot.ID, message string) error {
	if err := bot.tg.SendMessage(id, message); err != nil {
		return fmt.Errorf("при отправке сообщения %s произошла ошибка: %w", message, err)
	}

	return nil
}

func formatLinksMsg(links []tgbot.Link) string {
	if len(links) == 0 {
		return ""
	}

	builder := strings.Builder{}

	builder.WriteString("Список ваших ссылок:\n\n")

	for ind, link := range links {
		builder.WriteString(fmt.Sprintf("%d) %s", ind+1, link))
	}

	return builder.String()
}
