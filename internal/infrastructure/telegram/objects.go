package telegram

import "linkTraccer/internal/domain/tgbot"

type GetUpdateAnswer struct {
	DefaultServerAnswer
	Updates tgbot.Updates `json:"result"`
}

type DefaultServerAnswer struct {
	Ok bool `json:"ok"`
}

type SendMessage struct {
	ID   int64  `json:"chat_id"`
	Text string `json:"text"`
}
