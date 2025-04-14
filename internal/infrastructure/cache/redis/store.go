package redis

import (
	"fmt"
	"github.com/go-redis/redis"
	"linkTraccer/internal/domain/tgbot"
	"strconv"
)

type Store struct {
	client *redis.Client
}

func NewStore(config *Config) (*Store, error) {
	redisClient := redis.NewClient(&redis.Options{Addr: config.RedisAddr})

	if err := redisClient.Ping().Err(); err != nil {
		return nil, fmt.Errorf("ошибка при создании redis хранилища, не удалось установить соединение: %w", err)
	}

	return &Store{client: redisClient}, nil
}

func (s *Store) SetUserLinks(id tgbot.ID, links string) error {
	if err := s.client.Set(strconv.FormatInt(id, 10), links, 0).Err(); err != nil {
		return fmt.Errorf("ошибка при обновлении кеша пользователя c id %d: %w", id, err)
	}

	return nil
}

func (s *Store) GetUserLinks(id tgbot.ID) (string, error) {
	links, err := s.client.Get(strconv.FormatInt(id, 10)).Result()
	if err != nil {
		return links, fmt.Errorf("ошибка при получение ссылок пользователя из кеша: %w", err)
	}

	return links, err

}

func (s *Store) InvalidateUserCache(id tgbot.ID) error {
	if err := s.client.Del(strconv.FormatInt(id, 10)).Err(); err != nil {
		return fmt.Errorf("ошибка при инвалидации кеша, пользователя с id %d: %w", err)
	}

	return nil
}
