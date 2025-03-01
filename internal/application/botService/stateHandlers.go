package botService

import (
	"linkTraccer/internal/domain/tgbot"
	"log"
	"strings"
)

//stateHandler func(TgClient, tgbot.EventType)

func RegHandler(client TgClient, scrap ScrapClient, ctxStore CtxStorage, id ID, event tgbot.EventType) error {
	if err := ctxStore.RegUser(id); err != nil {
		return err
	}

	if err := scrap.RegUser(id); err != nil {
		return err
	}

	if err := client.SendMessage(id, FirstMessage); err != nil {
		return err
	}

	return nil
}

func CommandsHandler(client TgClient, scrap ScrapClient, ctxStore CtxStorage, id ID, event tgbot.EventType) error {
	switch event {
	case Start:
		if err := client.SendMessage(id, FirstMessage); err != nil {
			return err
		}
	case Help:
		if err := client.SendMessage(id, HelpMessage); err != nil {
			return err
		}
	case List:

		links, err := scrap.UserLinks(id)

		if err != nil {
			return err
		}

		if len(links) == 0 {
			if err = client.SendMessage(id, NoSavedLinks); err != nil {
				return err
			}
		} else {
			if err = client.SendMessage(id, strings.Join(links, "\n")); err != nil {
				return err
			}
		}

	case Untrack:
		if err := client.SendMessage(id, UntrackLink); err != nil {
			return err
		}

	case Track:
		if err := client.SendMessage(id, TrackLink); err != nil {
			return err
		}

	default:
		if err := client.SendMessage(id, UnknownCommand); err != nil {
			return err
		}
	}

	return nil
}

func LinkRemoveHandler(client TgClient, scrap ScrapClient, ctxStore CtxStorage, id ID, event tgbot.EventType) error {
	switch event {
	case Start:
		if err := client.SendMessage(id, FirstMessage); err != nil {
			return err
		}
	case Help:
		if err := client.SendMessage(id, HelpMessage); err != nil {
			return err
		}
	case List:

		links, err := scrap.UserLinks(id)

		if err != nil {
			return err
		}

		if len(links) == 0 {
			if err = client.SendMessage(id, NoSavedLinks); err != nil {
				return err
			}
		} else {
			if err = client.SendMessage(id, strings.Join(links, "\n")); err != nil {
				return err
			}
		}

	case Untrack:
		if err := client.SendMessage(id, UntrackLink); err != nil {
			return err
		}

	case Track:
		if err := client.SendMessage(id, TrackLink); err != nil {
			return err
		}

	default:
		if err := ctxStore.AddUrl(id, event); err != nil {
			return err
		}

		if err := scrap.RemoveLink(id, event); err != nil {

			if err := client.SendMessage(id, NotSaveThisLink); err != nil {
				return err
			}

		} else {
			if err := client.SendMessage(id, LinkDelete); err != nil {
				return err
			}
		}
	}
	return nil
}

func AddLinkHandler(client TgClient, scrap ScrapClient, ctxStore CtxStorage, id ID, event tgbot.EventType) error {

	switch event {
	case Start:
		if err := client.SendMessage(id, FirstMessage); err != nil {
			return err
		}
	case Help:
		if err := client.SendMessage(id, HelpMessage); err != nil {
			return err
		}
	case List:

		links, err := scrap.UserLinks(id)

		if err != nil {
			return err
		}

		if len(links) == 0 {
			if err = client.SendMessage(id, NoSavedLinks); err != nil {
				return err
			}
		} else {
			if err = client.SendMessage(id, strings.Join(links, "\n")); err != nil {
				return err
			}
		}

	case Untrack:
		if err := client.SendMessage(id, UntrackLink); err != nil {
			return err
		}

	case Track:
		if err := client.SendMessage(id, TrackLink); err != nil {
			return err
		}

	default:
		if err := ctxStore.AddUrl(id, event); err != nil {
			return err
		}

		if err := client.SendMessage(id, AddLinkTagMsg); err != nil {
			return err
		}
	}

	return nil
}

func AddTagHandler(client TgClient, scrap ScrapClient, ctxStore CtxStorage, id ID, event tgbot.EventType) error {

	switch event {
	case Start:
		if err := client.SendMessage(id, FirstMessage); err != nil {
			return err
		}
	case Help:
		if err := client.SendMessage(id, HelpMessage); err != nil {
			return err
		}
	case List:

		links, err := scrap.UserLinks(id)

		if err != nil {
			return err
		}

		if len(links) == 0 {
			if err = client.SendMessage(id, NoSavedLinks); err != nil {
				return err
			}
		} else {
			if err = client.SendMessage(id, strings.Join(links, "\n")); err != nil {
				return err
			}
		}

	case Untrack:
		if err := client.SendMessage(id, UntrackLink); err != nil {
			return err
		}

	case Track:
		if err := client.SendMessage(id, TrackLink); err != nil {
			return err
		}

	default:
		if err := ctxStore.AddTags(id, []string{event}); err != nil {
			return err
		}

		if err := client.SendMessage(id, AddLinkFilterMsg); err != nil {
			return err
		}

	}

	return nil
}

func SaveLinkHandler(client TgClient, scrap ScrapClient, ctxStore CtxStorage, id ID, event tgbot.EventType) error {

	switch event {
	case Start:
		if err := client.SendMessage(id, FirstMessage); err != nil {
			return err
		}
	case Help:
		if err := client.SendMessage(id, HelpMessage); err != nil {
			return err
		}
	case List:

		links, err := scrap.UserLinks(id)

		if err != nil {
			return err
		}

		if len(links) == 0 {
			if err = client.SendMessage(id, NoSavedLinks); err != nil {
				return err
			}
		} else {
			if err = client.SendMessage(id, strings.Join(links, "\n")); err != nil {
				return err
			}
		}

	case Untrack:
		if err := client.SendMessage(id, UntrackLink); err != nil {
			return err
		}

	case Track:
		if err := client.SendMessage(id, TrackLink); err != nil {
			return err
		}

	default:

		if err := ctxStore.AddFilters(id, []string{event}); err != nil {
			return err
		}

		userContext, err := ctxStore.UserContext(id)

		if err != nil {
			return err
		}

		ctxStore.ResetCtx(id)

		log.Println("Вызываю Add Link")

		if err := scrap.AddLink(id, userContext); err != nil {

			if err := client.SendMessage(id, WrongLink); err != nil {
				return err
			}

		} else {
			if err := client.SendMessage(id, GoodLink); err != nil {
				return err
			}
		}

	}

	return nil
}
