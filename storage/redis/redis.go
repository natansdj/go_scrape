package redis

import (
	"strconv"

	"github.com/natansdj/go_scrape/config"
	"github.com/natansdj/go_scrape/storage"

	"github.com/go-redis/redis/v7"
)

// New func implements the storage interface for go_scrape (https://github.com/natansdj/go_scrape)
func New(config config.ConfYaml) *Storage {
	return &Storage{
		config: config,
	}
}

func (s *Storage) getInt64(key string, count *int64) {
	val, _ := s.client.Get(key).Result()
	*count, _ = strconv.ParseInt(val, 10, 64)
}

// Storage is interface structure
type Storage struct {
	config config.ConfYaml
	client *redis.Client
}

// Init client storage.
func (s *Storage) Init() error {
	s.client = redis.NewClient(&redis.Options{
		Addr:     s.config.Stat.Redis.Addr,
		Password: s.config.Stat.Redis.Password,
		DB:       s.config.Stat.Redis.DB,
	})
	_, err := s.client.Ping().Result()

	return err
}

// Close the storage connection
func (s *Storage) Close() error {
	if s.client == nil {
		return nil
	}

	return s.client.Close()
}

// Reset Client storage.
func (s *Storage) Reset() {
	s.client.Set(storage.TotalCountKey, int64(0), 0)
}

// AddTotalCount record push notification count.
func (s *Storage) AddTotalCount(count int64) {
	s.client.IncrBy(storage.TotalCountKey, count)
}

// GetTotalCount show counts of all notification.
func (s *Storage) GetTotalCount() int64 {
	var count int64
	s.getInt64(storage.TotalCountKey, &count)

	return count
}
