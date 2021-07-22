package status

import (
	"github.com/natansdj/go_scrape/config"
	"github.com/natansdj/go_scrape/logx"
	"github.com/natansdj/go_scrape/storage"
	"github.com/natansdj/go_scrape/storage/memory"
	"github.com/natansdj/go_scrape/storage/redis"

	"github.com/thoas/stats"
)

// Stats provide response time, status code count, etc.
var Stats *stats.Stats

// StatStorage implements the storage interface
var StatStorage storage.Storage

// App is status structure
type App struct {
	Version    string `json:"version"`
	QueueMax   int    `json:"queue_max"`
	QueueUsage int    `json:"queue_usage"`
	TotalCount int64  `json:"total_count"`
}

// InitAppStatus for initialize app status
func InitAppStatus(conf config.ConfYaml) error {
	logx.LogAccess.Info("Init App Status Engine as ", conf.Stat.Engine)
	switch conf.Stat.Engine {
	case "redis":
		StatStorage = redis.New(conf)
	default:
		StatStorage = memory.New()
		//logx.LogError.Error("storage error: can't find storage driver")
		//return errors.New("can't find storage driver")
	}

	if err := StatStorage.Init(); err != nil {
		logx.LogError.Error("storage error: " + err.Error())

		return err
	}

	Stats = stats.New()

	return nil
}
