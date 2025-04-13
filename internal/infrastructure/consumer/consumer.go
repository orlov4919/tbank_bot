package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/segmentio/kafka-go"
	"linkTraccer/internal/application/botservice"
	"linkTraccer/internal/domain/dto"
	"log/slog"
	"strings"
)

func New(tgClient botservice.TgClient, config *Config, log *slog.Logger) *KafkaConsumer {
	return &KafkaConsumer{
		reader: kafka.NewReader(
			kafka.ReaderConfig{Brokers: strings.Split(config.Brokers, ","),
				Topic:    config.Topic,
				MaxBytes: 10e6}),
		tgClient: tgClient,
		log:      log,
	}
}

type KafkaConsumer struct {
	tgClient botservice.TgClient
	log      *slog.Logger
	reader   *kafka.Reader
}

func (c *KafkaConsumer) ReadUserUpdates(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			c.log.Info("consumer завершил свою работу")
			return nil

		default:
			msg, err := c.reader.ReadMessage(context.Background())

			if err != nil && errors.Is(err, ctx.Err()) {
				c.log.Info("consumer завершил свою работу")
				return nil
			}

			if err != nil {
				c.log.Error("ошибка при получении обновлений пользователей из топика", "err", err)
				continue
			}

			updates := &dto.LinkUpdate{}

			if err := json.Unmarshal(msg.Value, updates); err != nil {
				c.log.Error("ошибка при анмаршалинге сообщений из топика", "err", err)
				continue
			}

			if err := c.processUpdate(updates); err != nil {
				c.log.Error("ошибка при обработке обновлений", "err", err.Error())
			}
		}
	}
}

func (c *KafkaConsumer) processUpdate(updates *dto.LinkUpdate) error {
	msg := updates.Description + updates.URL

	for _, userID := range updates.TgChatIDs {
		if err := c.tgClient.SendMessage(userID, msg); err != nil {
			return fmt.Errorf("ошибка при отправке обновлений в телеграмм: %w", err)
		}
	}

	return nil
}

func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}
