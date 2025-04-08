package botservice

import (
	"fmt"
	"linkTraccer/internal/domain/tgbot"
	"strings"
)

func RegHandler(client TgClient, scrap ScrapClient, ctxStore CtxStorage, id ID, _ tgbot.EventType) error {
	if err := ctxStore.RegUser(id); err != nil {
		return fmt.Errorf("при регистрации в хранилище контекстной информации возникла ошибка: %w", err)
	}

	if err := scrap.RegUser(id); err != nil {
		return fmt.Errorf("при регистрации пользователя произошла ошибка: %w", err)
	}

	if err := client.SendMessage(id, FirstMessage); err != nil {
		return fmt.Errorf("при отправке %s произошла ошибка: %w", FirstMessage, err)
	}

	return nil
}

func CommandsStateHandler(client TgClient, scrap ScrapClient, ctxStore CtxStorage, id ID, event tgbot.EventType) error {
	err := CommandsHandler(client, scrap, ctxStore, id, event)

	if err != nil {
		if err := client.SendMessage(id, UnknownCommand); err != nil {
			return fmt.Errorf("при отправке сообщения %s произошла ошибка: %w", UnknownCommand, err)
		}
	}

	return nil
}

func LinkRemoveHandler(client TgClient, scrap ScrapClient, ctxStore CtxStorage, id ID, event tgbot.EventType) error {
	err := CommandsHandler(client, scrap, ctxStore, id, event)
	if err == nil {
		return nil
	}

	if err := ctxStore.AddURL(id, event); err != nil {
		return fmt.Errorf("при добавлении ссылки в контекстное хранилище произошла ошибка: %w", err)
	}

	if err := scrap.RemoveLink(id, event); err != nil {
		return sendMessageWithError(client, id, NotSaveThisLink)
	}

	return sendMessageWithError(client, id, LinkDelete)
}

func sendMessageWithError(client TgClient, id ID, message string) error {
	errorFormat := "при отправке сообщения %s произошла ошибка: %w"
	if err := client.SendMessage(id, message); err != nil {
		return fmt.Errorf(errorFormat, message, err)
	}

	return nil
}

func AddLinkHandler(client TgClient, scrap ScrapClient, ctxStore CtxStorage, id ID, event tgbot.EventType) error {
	err := CommandsHandler(client, scrap, ctxStore, id, event)

	if err != nil {
		if err := ctxStore.AddURL(id, event); err != nil {
			return fmt.Errorf("при добавлении ссылки в контекстное хранилище, произошла ошибка :%w", err)
		}

		if err := client.SendMessage(id, AddLinkTagMsg); err != nil {
			return fmt.Errorf("при отправке сообщения %s произошла ошибка: %w", AddLinkTagMsg, err)
		}
	}

	return nil
}

func AddTagHandler(client TgClient, scrap ScrapClient, ctxStore CtxStorage, id ID, event tgbot.EventType) error {
	err := CommandsHandler(client, scrap, ctxStore, id, event)

	if err != nil {
		if err := ctxStore.AddTags(id, []string{event}); err != nil { // временное решение
			return fmt.Errorf("при добавлении тегов в контекстное хранилище, произошла ошибка :%w", err)
		}

		if err := client.SendMessage(id, AddLinkFilterMsg); err != nil {
			return fmt.Errorf("при отправке сообщения %s произошла ошибка: %w", AddLinkFilterMsg, err)
		}
	}

	return nil
}

func SaveLinkHandler(client TgClient, scrap ScrapClient, ctxStore CtxStorage, id ID, event tgbot.EventType) error {
	err := CommandsHandler(client, scrap, ctxStore, id, event)
	if err == nil {
		return nil
	}

	if err := ctxStore.AddFilters(id, []string{event}); err != nil {
		return fmt.Errorf("при добавлении фильтров в контекстное хранилище произошла ошибка: %w", err)
	}

	userContext, err := ctxStore.UserContext(id)
	if err != nil {
		return fmt.Errorf("при получении контекстной информации произошла ошибка: %w", err)
	}

	if err := scrap.AddLink(id, userContext); err != nil {
		return sendMessageWithError(client, id, WrongLink)
	}

	return sendMessageWithError(client, id, GoodLink)
}

func CommandsHandler(client TgClient, scrap ScrapClient, _ CtxStorage, id ID, event tgbot.EventType) error {
	switch event {
	case Start:
		if err := client.SendMessage(id, FirstMessage); err != nil {
			return fmt.Errorf("при отправке сообщения %s произошла ошибка: %w", FirstMessage, err)
		}
	case Help:
		if err := client.SendMessage(id, HelpMessage); err != nil {
			return fmt.Errorf("при отправке сообщения %s произошла ошибка: %w", HelpMessage, err)
		}
	case List:
		links, err := scrap.UserLinks(id)

		if err != nil {
			return fmt.Errorf("при получении ссылок произошла ошибка: %w", err)
		}

		if len(links) == 0 {
			if err = client.SendMessage(id, NoSavedLinks); err != nil {
				return fmt.Errorf("при отправке сообщения %s произошла ошибка: %w", NoSavedLinks, err)
			}
		} else {
			if err = client.SendMessage(id, formatLinksMsg(links)); err != nil {
				return fmt.Errorf("при отправке ссылок произошла ошибка: %w", err)
			}
		}

	case Untrack:
		if err := client.SendMessage(id, UntrackLink); err != nil {
			return fmt.Errorf("при отправке сообщения %s произошла ошибка: %w", UntrackLink, err)
		}

	case Track:
		if err := client.SendMessage(id, TrackLink); err != nil {
			return fmt.Errorf("при отправке сообщения %s произошла ошибка: %w", TrackLink, err)
		}
	default:
		return NewErrCommandNotFound(event)
	}

	return nil
}

func formatLinksMsg(links []Link) string {
	builder := strings.Builder{}

	builder.WriteString("Список ваших ссылок:\n\n")

	for ind, link := range links {
		builder.WriteString(fmt.Sprintf("%d) %s", ind+1, link))
	}

	return builder.String()
}
