package tgbot

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

type Updates = []Update

type Link = string
