package config

import (
	"bytes"
	"errors"
	"github.com/natansdj/go_scrape/logx"
	"io/ioutil"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

// ConfYaml is config structure.
type ConfYaml struct {
	Core   SectionCore  `yaml:"core"`
	API    SectionAPI   `yaml:"api"`
	Source SourceAPI    `yaml:"source"`
	DB     DatabaseAPI  `yaml:"database"`
	Log    SectionLog   `yaml:"log"`
	Queue  SectionQueue `yaml:"queue"`
	Stat   SectionStat  `yaml:"stat"`
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

// SourceAPI
type SourceAPI struct {
	BaseURI               string `yaml:"base_uri"`
	CtxTimeout            int    `yaml:"ctx_timeout"`
	CtxKeepAlive          int    `yaml:"ctx_keepalive"`
	MaxIdleConnsPerHost   int    `yaml:"max_idle_cons_per_host"`
	MaxIdleConns          int    `yaml:"max_idle_con"`
	IdleConnTimeout       int    `yaml:"idle_con_timeout"`
	TLSHandshakeTimeout   int    `yaml:"tls_handshake_timeout"`
	ExpectContinueTimeout int    `yaml:"expect_continue_timeout"`
	HttpTimeout           int    `yaml:"http_timeout"`
}

// DatabaseAPI
type DatabaseAPI struct {
	Host     string `yaml:"db_host"`
	Port     string `yaml:"db_port"`
	User     string `yaml:"db_user"`
	Password string `yaml:"db_password"`
	Name     string `yaml:"db_name"`
	Charset  string `yaml:"db_charset"`
	Prefix   string `yaml:"db_prefix"`
	Timezone string `yaml:"db_timezone"`
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
			defer logx.LogAccess.Info("Using config file:", viper.ConfigFileUsed())
		} else {
			errMsg := "config file not found! "
			logx.LogError.Fatal(errMsg)
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

	// Source
	conf.Source.BaseURI = viper.GetString("source.base_uri")
	conf.Source.CtxTimeout = viper.GetInt("source.ctx_timeout")
	conf.Source.CtxKeepAlive = viper.GetInt("source.ctx_keepalive")
	conf.Source.MaxIdleConnsPerHost = viper.GetInt("source.max_idle_cons_per_host")
	conf.Source.MaxIdleConns = viper.GetInt("source.max_idle_con")
	conf.Source.IdleConnTimeout = viper.GetInt("source.idle_con_timeout")
	conf.Source.TLSHandshakeTimeout = viper.GetInt("source.tls_handshake_timeout")
	conf.Source.ExpectContinueTimeout = viper.GetInt("source.expect_continue_timeout")
	conf.Source.HttpTimeout = viper.GetInt("source.http_timeout")

	// Database
	conf.DB.Host = viper.GetString("database.db_host")
	conf.DB.Port = viper.GetString("database.db_port")
	conf.DB.User = viper.GetString("database.db_user")
	conf.DB.Password = viper.GetString("database.db_password")
	conf.DB.Name = viper.GetString("database.db_name")
	conf.DB.Charset = viper.GetString("database.db_charset")
	conf.DB.Prefix = viper.GetString("database.db_prefix")
	conf.DB.Timezone = viper.GetString("database.db_timezone")

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
