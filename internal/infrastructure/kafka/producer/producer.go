package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"linkTraccer/internal/domain/dto"
	"strings"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

func New(config *Config) *KafkaProducer {
	return &KafkaProducer{writer: &kafka.Writer{
		Addr:      kafka.TCP(strings.Split(config.Brokers, ",")...),
		Topic:     config.Topic,
		BatchSize: config.Bath,
		Balancer:  &kafka.RoundRobin{},
	}}
}

func (k *KafkaProducer) SendLinkUpdates(update *dto.LinkUpdate) error {
	updatesJSON, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("ошибка при маршилинге апдейта в кафка продюсере: %w", err)
	}

	err = k.writer.WriteMessages(context.Background(), kafka.Message{Value: updatesJSON})
	if err != nil {
		return fmt.Errorf("ошибка при отправке обновлений в топик: %w", err)
	}

	return nil
}

func (k *KafkaProducer) Close() error {
	return k.writer.Close()
}
