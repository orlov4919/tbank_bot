package main

import (
	"dz/internal/domain/tgbot"
	"log"
	"time"

	"github.com/BurntSushi/toml"
)

const (
	pathBotConfig = "../../configs/bot/config.toml"
	host          = "api.telegram.org"
	errChanSize   = 10000
)

func main() {
	botConfig := tgbot.NewConfig()

	if _, err := toml.DecodeFile(pathBotConfig, botConfig); err != nil {
		log.Fatal(err)
	}

	bot := tgbot.New(botConfig, host)
	errChan := make(chan error, errChanSize)

	defer close(errChan)

	go func() {
		for err := range errChan {
			log.Println(err)
		}
	}()

	for {
		updates, err := bot.HandleUsersUpdates()

		if err != nil {
			errChan <- err
			continue
		}

		if len(updates) > 0 {
			bot.ChangeOffset(updates[len(updates)-1].UpdateID + 1)

			log.Printf("Получено %d новых апдейтов", len(updates))

			for _, update := range updates {
				go func(update tgbot.Update) {
					err := bot.SendMessage(update.Msg.From.ID, update.Msg.Text)

					if err != nil {
						errChan <- err
					}
				}(update)
			}
		}

		time.Sleep(time.Second * 5)
	}
}
