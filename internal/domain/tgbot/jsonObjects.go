package tgbot

type SendMessage struct {
	ID   int    `json:"chat_id"`
	Text string `json:"text"`
}

type Update struct {
	UpdateID int     `json:"update_id"`
	Msg      Message `json:"message"`
}

type Message struct {
	From User   `json:"from"`
	Text string `json:"text"`
}

type User struct {
	ID int `json:"id"`
}

type GetUpdateAnswer struct {
	DefaultServerAnswer
	Updates Updates `json:"result"`
}

type DefaultServerAnswer struct {
	Ok bool `json:"ok"`
}

type Updates = []Update
