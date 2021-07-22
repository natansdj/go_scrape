package redis

import (
	"sync"
	"testing"

	"github.com/natansdj/go_scrape/config"
	"github.com/stretchr/testify/assert"
)

func TestRedisServerError(t *testing.T) {
	cfg, _ := config.LoadConf()
	cfg.Stat.Redis.Addr = "redis:6370"

	redis := New(cfg)
	err := redis.Init()

	assert.Error(t, err)
}

func TestRedisEngine(t *testing.T) {
	var val int64

	cfg, _ := config.LoadConf()
	cfg.Stat.Redis.Addr = "redis:6379"

	redis := New(cfg)
	err := redis.Init()
	assert.Nil(t, err)
	redis.Reset()

	redis.AddTotalCount(10)
	val = redis.GetTotalCount()
	assert.Equal(t, int64(10), val)
	redis.AddTotalCount(10)
	val = redis.GetTotalCount()
	assert.Equal(t, int64(20), val)

	// test concurrency issues
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			redis.AddTotalCount(1)
			wg.Done()
		}()
	}
	wg.Wait()
	val = redis.GetTotalCount()
	assert.Equal(t, int64(10), val)

	assert.NoError(t, redis.Close())
}
