package status

import (
	"testing"
	"time"

	"github.com/natansdj/go_scrape/config"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	m.Run()
}

func TestStorageDriverExist(t *testing.T) {
	cfg, _ := config.LoadConf()
	cfg.Stat.Engine = "Test"
	err := InitAppStatus(cfg)
	assert.Error(t, err)
}

func TestStatForMemoryEngine(t *testing.T) {
	// wait android push notification response.
	time.Sleep(5 * time.Second)

	var val int64
	cfg, _ := config.LoadConf()
	cfg.Stat.Engine = "memory"
	err := InitAppStatus(cfg)
	assert.Nil(t, err)

	StatStorage.AddTotalCount(100)

	val = StatStorage.GetTotalCount()
	assert.Equal(t, int64(100), val)
}

func TestRedisServerSuccess(t *testing.T) {
	cfg, _ := config.LoadConf()
	cfg.Stat.Engine = "redis"
	cfg.Stat.Redis.Addr = "redis:6379"

	err := InitAppStatus(cfg)

	assert.NoError(t, err)
}

func TestRedisServerError(t *testing.T) {
	cfg, _ := config.LoadConf()
	cfg.Stat.Engine = "redis"
	cfg.Stat.Redis.Addr = "redis:6370"

	err := InitAppStatus(cfg)

	assert.Error(t, err)
}

func TestStatForRedisEngine(t *testing.T) {
	var val int64
	cfg, _ := config.LoadConf()
	cfg.Stat.Engine = "redis"
	cfg.Stat.Redis.Addr = "redis:6379"
	err := InitAppStatus(cfg)
	assert.Nil(t, err)

	StatStorage.Init()
	StatStorage.Reset()

	StatStorage.AddTotalCount(100)

	val = StatStorage.GetTotalCount()
	assert.Equal(t, int64(100), val)
}

func TestDefaultEngine(t *testing.T) {
	var val int64
	// defaul engine as memory
	cfg, _ := config.LoadConf()
	err := InitAppStatus(cfg)
	assert.Nil(t, err)

	StatStorage.Reset()

	StatStorage.AddTotalCount(100)

	val = StatStorage.GetTotalCount()
	assert.Equal(t, int64(100), val)
}
