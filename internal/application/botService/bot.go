package botService

import (
	"linkTraccer/internal/domain/tgbot"
	"log"
	"time"
)

type Updates = tgbot.Updates

//type Config struct {
//	Token string `toml:"token"`
//}
//
//func NewConfig() *Config {
//	return &Config{}
//}

type TgClient interface {
	HandleUsersUpdates(offset, limit int) (Updates, error)
	SendMessage(userID int, text string) error
}

type TgBot struct {
	offset int
	limit  int
	client TgClient
}

func New(client TgClient, limit int) *TgBot {
	return &TgBot{
		client: client,
		limit:  limit,
	}
}

func (bot *TgBot) Start() {

	for {
		updates, _ := bot.client.HandleUsersUpdates(bot.offset, bot.limit) // надо подумать что делать с ошибкой
		updatesNum := len(updates)

		if updatesNum > 0 {
			bot.changeOffset(updates[updatesNum-1].UpdateID + 1)

			log.Printf("Получено %d новых апдейтов", len(updates))

			for _, update := range updates {
				go func(update tgbot.Update) {
					// err := bot.SendMessage(update.Msg.From.ID, update.Msg.Text)

				}(update)
			}
		}

		time.Sleep(time.Second * 5)

	}

}

func (bot *TgBot) changeOffset(newOffset int) {
	bot.offset = newOffset
}
