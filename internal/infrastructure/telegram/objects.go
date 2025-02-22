package telegram

type GetUpdateAnswer struct {
	DefaultServerAnswer
	Updates Updates `json:"result"`
}

type DefaultServerAnswer struct {
	Ok bool `json:"ok"`
}

type SendMessage struct {
	ID   int    `json:"chat_id"`
	Text string `json:"text"`
}
