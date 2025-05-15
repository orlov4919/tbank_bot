package botservice

import "linkTraccer/internal/domain/tgbot"

const (
	InitialState         tgbot.State = "init"    // В этом состоянии бот может принять только команду /start
	AnyRegisteredCommand tgbot.State = "comands" // В этом состоянии бот может выполнить любую команду
	RemoveLink           tgbot.State = "remove"  // В этом состоянии бот ждет ссылку для удаления, а так же может выполнить любую команду
	AddNewLink           tgbot.State = "link"    // В этом состоянии бот ждет ссылку, а так же может выполнить любую команду
	AddLinkTag           tgbot.State = "tag"     // В этом состоянии бот ждет тэг ссылки, а так же может выполнить любую команду
	AddLinkFilter        tgbot.State = "filter"  // В этом состоянии бот ждет фильтр ссылки, а так же может выполнить любую команду
)

func NewTransition(event tgbot.Event, dst tgbot.State) tgbot.Transition {
	return tgbot.Transition{
		Event: event,
		Dst:   dst,
	}
}

var (
	StartTransition   = NewTransition(Start, AnyRegisteredCommand)
	HelpTransition    = NewTransition(Help, AnyRegisteredCommand)
	UntrackTransition = NewTransition(Untrack, RemoveLink)
	RemoveTransition  = NewTransition(tgbot.TextEvent, AnyRegisteredCommand)
	ListTransition    = NewTransition(List, AnyRegisteredCommand)
	TrackTransition   = NewTransition(Track, AddNewLink)
	LinkTransition    = NewTransition(tgbot.TextEvent, AddLinkTag)
	TagTransition     = NewTransition(tgbot.TextEvent, AddLinkFilter)
	FilterTransition  = NewTransition(tgbot.TextEvent, AnyRegisteredCommand)
)

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

func botStates() tgbot.States {
	return states
}
