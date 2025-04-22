package producer_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	kafkatest "github.com/testcontainers/testcontainers-go/modules/kafka"
	"linkTraccer/internal/domain/dto"
	producer2 "linkTraccer/internal/infrastructure/kafka/producer"
	"strings"
	"testing"
)

const (
	topic    = "updates"
	bathSize = 1
)

func SetupContainer(ctx context.Context, t *testing.T) *kafkatest.KafkaContainer {
	kafkaContainer, err := kafkatest.Run(ctx,
		"confluentinc/confluent-local:7.5.0")

	assert.NoError(t, err)

	err = CreateTopic(ctx, kafkaContainer)

	assert.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, testcontainers.TerminateContainer(kafkaContainer))
	})

	return kafkaContainer
}

func CreateTopic(ctx context.Context, container *kafkatest.KafkaContainer) error {
	cmd := []string{
		"/bin/bash",
		"-c",
		fmt.Sprintf(
			"/usr/bin/kafka-topics --bootstrap-server localhost:9092 --create --topic %s --partitions 1 --replication-factor 1",
			topic,
		),
	}

	_, _, err := container.Exec(ctx, cmd)

	return err
}

func TestKafkaProducer_SendLinkUpdates(t *testing.T) {
	kafkaContainer := SetupContainer(context.Background(), t)
	brokersSlice, err := kafkaContainer.Brokers(context.Background())

	assert.NoError(t, err)

	brokers := strings.Join(brokersSlice, ",")

	producer := producer2.New(&producer2.Config{Brokers: brokers, Bath: bathSize, Topic: topic})
	consumer := kafka.NewReader(kafka.ReaderConfig{Brokers: brokersSlice, Topic: topic})

	defer producer.Close()
	defer consumer.Close()

	type TestCase struct {
		update *dto.LinkUpdate
	}

	tests := []TestCase{
		{
			update: &dto.LinkUpdate{
				ID:          1,
				URL:         "http://localhost:8080",
				Description: "Новое обновление",
			},
		},
		{
			update: &dto.LinkUpdate{
				ID:          2,
				URL:         "http://localhost:8080",
				Description: "Новое обновление",
				TgChatIDs:   []int64{1, 2, 3},
			},
		},
	}

	for _, test := range tests {
		err := producer.SendLinkUpdates(test.update)

		assert.NoError(t, err)

		msg, err := consumer.ReadMessage(context.Background())

		assert.NoError(t, err)

		getUpdates := &dto.LinkUpdate{}

		err = json.Unmarshal(msg.Value, getUpdates)

		assert.NoError(t, err)
		assert.Equal(t, test.update, getUpdates)
	}
}
