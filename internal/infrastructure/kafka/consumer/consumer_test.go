package consumer_test

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	kafkatest "github.com/testcontainers/testcontainers-go/modules/kafka"
	"linkTraccer/internal/domain/dto"
	consumer2 "linkTraccer/internal/infrastructure/kafka/consumer"
	"linkTraccer/internal/infrastructure/kafka/consumer/mocks"
	producer2 "linkTraccer/internal/infrastructure/kafka/producer"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	topic    = "updates"
	bathSize = 1
	firstID  = int64(1)
	secondID = int64(2)
)

var (
	logLevel = slog.LevelDebug
	logger   = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))

	usersUpdate = &dto.LinkUpdate{
		ID:          2,
		URL:         "http://localhost:8080",
		Description: "Новое обновление",
		TgChatIDs:   []int64{firstID, secondID},
	}

	msg = usersUpdate.Description + usersUpdate.URL
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

func TestKafkaConsumer_ReadUserUpdates(t *testing.T) {
	kafkaContainer := SetupContainer(context.Background(), t)
	brokersSlice, err := kafkaContainer.Brokers(context.Background())

	assert.NoError(t, err)

	brokers := strings.Join(brokersSlice, ",")

	tgClient := mocks.NewTgClient(t)

	tgClient.On("SendMessage", firstID, msg).Return(nil).Once()
	tgClient.On("SendMessage", secondID, msg).Return(nil).Once()

	producer := producer2.New(&producer2.Config{Brokers: brokers, Bath: bathSize, Topic: topic})
	consumer := consumer2.New(tgClient, &consumer2.Config{Brokers: brokers, Topic: topic, Batch: bathSize}, logger)
	ctx, cancel := context.WithCancel(context.Background())

	go consumer.ReadUserUpdates(ctx)

	defer producer.Close()
	defer consumer.Close()

	type TestCase struct {
		update *dto.LinkUpdate
	}

	tests := []TestCase{
		{
			update: usersUpdate,
		},
	}

	for _, test := range tests {
		err := producer.SendLinkUpdates(test.update)

		assert.NoError(t, err)
	}
	// ожидаем окончания выполнения горутин
	time.Sleep(time.Second * 2)
	cancel()
	tgClient.AssertExpectations(t)
}
