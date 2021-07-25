package config

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

// ConfYaml is config structure.
type ConfYaml struct {
	Core  SectionCore  `yaml:"core"`
	API   SectionAPI   `yaml:"api"`
	Log   SectionLog   `yaml:"log"`
	Queue SectionQueue `yaml:"queue"`
	Stat  SectionStat  `yaml:"stat"`
}

// SectionCore is sub section of config.
type SectionCore struct {
	Enabled         bool           `yaml:"enabled"`
	Address         string         `yaml:"address"`
	ShutdownTimeout int64          `yaml:"shutdown_timeout"`
	Port            string         `yaml:"port"`
	MaxNotification int64          `yaml:"max_notification"`
	WorkerNum       int64          `yaml:"worker_num"`
	QueueNum        int64          `yaml:"queue_num"`
	Mode            string         `yaml:"mode"`
	Sync            bool           `yaml:"sync"`
	SSL             bool           `yaml:"ssl"`
	CertPath        string         `yaml:"cert_path"`
	KeyPath         string         `yaml:"key_path"`
	CertBase64      string         `yaml:"cert_base64"`
	KeyBase64       string         `yaml:"key_base64"`
	HTTPProxy       string         `yaml:"http_proxy"`
	FeedbackURL     string         `yaml:"feedback_hook_url"`
	FeedbackTimeout int64          `yaml:"feedback_timeout"`
	PID             SectionPID     `yaml:"pid"`
	AutoTLS         SectionAutoTLS `yaml:"auto_tls"`
}

// SectionAPI is sub section of config.
type SectionAPI struct {
	PushURI    string `yaml:"push_uri"`
	StatGoURI  string `yaml:"stat_go_uri"`
	StatAppURI string `yaml:"stat_app_uri"`
	ConfigURI  string `yaml:"config_uri"`
	SysStatURI string `yaml:"sys_stat_uri"`
	MetricURI  string `yaml:"metric_uri"`
	HealthURI  string `yaml:"health_uri"`
}

// SectionLog is sub section of config.
type SectionLog struct {
	Format      string `yaml:"format"`
	AccessLog   string `yaml:"access_log"`
	AccessLevel string `yaml:"access_level"`
	ErrorLog    string `yaml:"error_log"`
	ErrorLevel  string `yaml:"error_level"`
	HideToken   bool   `yaml:"hide_token"`
}

// SectionStat is sub section of config.
type SectionStat struct {
	Engine string       `yaml:"engine"`
	Redis  SectionRedis `yaml:"redis"`
}

// SectionQueue is sub section of config.
type SectionQueue struct {
	Engine string     `yaml:"engine"`
	NSQ    SectionNSQ `yaml:"nsq"`
}

// SectionNSQ is sub section of config.
type SectionNSQ struct {
	Addr    string `yaml:"addr"`
	Topic   string `yaml:"topic"`
	Channel string `yaml:"channel"`
}

// SectionRedis is sub section of config.
type SectionRedis struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// SectionAutoTLS support Let's Encrypt setting.
type SectionAutoTLS struct {
	Enabled bool   `yaml:"enabled"`
	Folder  string `yaml:"folder"`
	Host    string `yaml:"host"`
}

// SectionPID is sub section of config.
type SectionPID struct {
	Enabled  bool   `yaml:"enabled"`
	Path     string `yaml:"path"`
	Override bool   `yaml:"override"`
}

// LoadConf load config from file and read in environment variables that match
func LoadConf(confPath ...string) (ConfYaml, error) {
	var conf ConfYaml

	viper.SetConfigType("yaml")
	viper.AutomaticEnv()            // read in environment variables that match
	viper.SetEnvPrefix("go_scrape") // will be uppercased automatically
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if len(confPath) > 0 && confPath[0] != "" {
		content, err := ioutil.ReadFile(confPath[0])
		if err != nil {
			return conf, err
		}

		if err := viper.ReadConfig(bytes.NewBuffer(content)); err != nil {
			return conf, err
		}
	} else {
		// Search config in home directory with name ".go_scrape" (without extension).
		viper.AddConfigPath("/config/")
		viper.AddConfigPath("$HOME/.go_scrape")
		viper.AddConfigPath(".")
		viper.SetConfigName("config")

		// If a config file is found, read it in.
		if err := viper.ReadInConfig(); err == nil {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		} else {
			errMsg := "config file not found! "
			log.Fatal(errMsg)
			return conf, errors.New(errMsg)
		}
	}

	// Core
	conf.Core.Address = viper.GetString("core.address")
	conf.Core.Port = viper.GetString("core.port")
	conf.Core.ShutdownTimeout = int64(viper.GetInt("core.shutdown_timeout"))
	conf.Core.Enabled = viper.GetBool("core.enabled")
	conf.Core.WorkerNum = int64(viper.GetInt("core.worker_num"))
	conf.Core.QueueNum = int64(viper.GetInt("core.queue_num"))
	conf.Core.Mode = viper.GetString("core.mode")
	conf.Core.Sync = viper.GetBool("core.sync")
	conf.Core.FeedbackURL = viper.GetString("core.feedback_hook_url")
	conf.Core.FeedbackTimeout = int64(viper.GetInt("core.feedback_timeout"))
	conf.Core.SSL = viper.GetBool("core.ssl")
	conf.Core.CertPath = viper.GetString("core.cert_path")
	conf.Core.KeyPath = viper.GetString("core.key_path")
	conf.Core.CertBase64 = viper.GetString("core.cert_base64")
	conf.Core.KeyBase64 = viper.GetString("core.key_base64")
	conf.Core.MaxNotification = int64(viper.GetInt("core.max_notification"))
	conf.Core.HTTPProxy = viper.GetString("core.http_proxy")
	conf.Core.PID.Enabled = viper.GetBool("core.pid.enabled")
	conf.Core.PID.Path = viper.GetString("core.pid.path")
	conf.Core.PID.Override = viper.GetBool("core.pid.override")
	conf.Core.AutoTLS.Enabled = viper.GetBool("core.auto_tls.enabled")
	conf.Core.AutoTLS.Folder = viper.GetString("core.auto_tls.folder")
	conf.Core.AutoTLS.Host = viper.GetString("core.auto_tls.host")

	// Api
	conf.API.PushURI = viper.GetString("api.push_uri")
	conf.API.StatGoURI = viper.GetString("api.stat_go_uri")
	conf.API.StatAppURI = viper.GetString("api.stat_app_uri")
	conf.API.ConfigURI = viper.GetString("api.config_uri")
	conf.API.SysStatURI = viper.GetString("api.sys_stat_uri")
	conf.API.MetricURI = viper.GetString("api.metric_uri")
	conf.API.HealthURI = viper.GetString("api.health_uri")

	// log
	conf.Log.Format = viper.GetString("log.format")
	conf.Log.AccessLog = viper.GetString("log.access_log")
	conf.Log.AccessLevel = viper.GetString("log.access_level")
	conf.Log.ErrorLog = viper.GetString("log.error_log")
	conf.Log.ErrorLevel = viper.GetString("log.error_level")
	conf.Log.HideToken = viper.GetBool("log.hide_token")

	// Queue Engine
	conf.Queue.Engine = viper.GetString("queue.engine")
	conf.Queue.NSQ.Addr = viper.GetString("queue.nsq.addr")
	conf.Queue.NSQ.Topic = viper.GetString("queue.nsq.topic")
	conf.Queue.NSQ.Channel = viper.GetString("queue.nsq.channel")

	// Stat Engine
	conf.Stat.Engine = viper.GetString("stat.engine")
	conf.Stat.Redis.Addr = viper.GetString("stat.redis.addr")
	conf.Stat.Redis.Password = viper.GetString("stat.redis.password")
	conf.Stat.Redis.DB = viper.GetInt("stat.redis.db")

	if conf.Core.WorkerNum == int64(0) {
		conf.Core.WorkerNum = int64(runtime.NumCPU())
	}

	if conf.Core.QueueNum == int64(0) {
		conf.Core.QueueNum = int64(8192)
	}

	return conf, nil
}
