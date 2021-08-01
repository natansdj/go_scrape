package router

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/natansdj/go_scrape/metric"
	"github.com/natansdj/go_scrape/models"
	"github.com/natansdj/go_scrape/status"
	"net/http"
	"os"
	"sync"

	"github.com/natansdj/go_scrape/config"
	"github.com/natansdj/go_scrape/core"
	"github.com/natansdj/go_scrape/go_scrape"
	"github.com/natansdj/go_scrape/logx"
	"github.com/natansdj/go_scrape/queue"

	api "github.com/appleboy/gin-status-api"
	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/mattn/go-isatty"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/thoas/stats"
	"golang.org/x/crypto/acme/autocert"
)

var (
	isTerm bool
	doOnce sync.Once
)

func init() {
	isTerm = isatty.IsTerminal(os.Stdout.Fd())
}

func abortWithError(c *gin.Context, code int, message string) {
	c.AbortWithStatusJSON(code, gin.H{
		"code":    code,
		"message": message,
	})
}

func rootHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"text": "Welcome to go_scrape server.",
	})
}

func heartbeatHandler(c *gin.Context) {
	c.AbortWithStatus(http.StatusOK)
}

func versionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"source":  "https://github.com/natansdj/go_scrape",
		"version": GetVersion(),
	})
}

func debugPushHandler(c *gin.Context) {
	logx.LogPush(&logx.InputLog{
		ID:        "",
		Status:    core.SucceededPush,
		Token:     "",
		Message:   "DEBUG",
		Platform:  0,
		Error:     nil,
		HideToken: false,
		Format:    "json",
	})
	c.JSON(http.StatusOK, gin.H{
		"text": "DEBUG Pushed",
	})
}

func pushHandler(q *queue.Queue) gin.HandlerFunc {
	cfg := models.CFG
	return func(c *gin.Context) {
		var form go_scrape.RequestPush
		var msg string

		if err := c.ShouldBindWith(&form, binding.JSON); err != nil {
			msg = "Missing notifications field."
			logx.LogAccess.Debug(err)
			abortWithError(c, http.StatusBadRequest, msg)
			return
		}

		if len(form.Notifications) == 0 {
			msg = "Notifications field is empty."
			logx.LogAccess.Debug(msg)
			abortWithError(c, http.StatusBadRequest, msg)
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			// Deprecated: the CloseNotifier interface predates Go's context package.
			// New code should use Request.Context instead.
			// Change to context package
			<-c.Request.Context().Done()
			// Don't send notification after client timeout or disconnected.
			// See the following issue for detail information.
			// https://github.com/natansdj/go_scrape/issues/422
			if cfg.Core.Sync {
				cancel()
			}
		}()

		counts, logs := handleNotification(ctx, cfg, form, q)

		c.JSON(http.StatusOK, gin.H{
			"success": "ok",
			"counts":  counts,
			"logs":    logs,
		})
	}
}

func configHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.YAML(http.StatusCreated, models.CFG)
	}
}

func metricsHandler(c *gin.Context) {
	logx.LogAccess.Debug("metricsHandler")
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}

func appStatusHandler(q *queue.Queue) gin.HandlerFunc {
	return func(c *gin.Context) {
		result := status.App{}

		result.Version = GetVersion()
		result.QueueMax = q.Capacity()
		result.QueueUsage = q.Usage()
		result.TotalCount = status.StatStorage.GetTotalCount()

		c.JSON(http.StatusOK, result)
	}
}

func sysStatsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, status.Stats.Data())
	}
}

// StatMiddleware response time, status code count, etc.
func StatMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		beginning, recorder := status.Stats.Begin(c.Writer)
		c.Next()
		status.Stats.End(beginning, stats.WithRecorder(recorder))
	}
}

func autoTLSServer(cfg config.ConfYaml, q *queue.Queue) *http.Server {
	m := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(cfg.Core.AutoTLS.Host),
		Cache:      autocert.DirCache(cfg.Core.AutoTLS.Folder),
	}

	return &http.Server{
		Addr:      ":https",
		TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
		Handler:   routerEngine(cfg, q),
	}
}

// markFailedNotification adds failure logs for all tokens in push notification
func markFailedNotification(cfg config.ConfYaml, notification *go_scrape.PushNotification, reason string) {
	logx.LogError.Error(reason)
	for _, token := range notification.Tokens {
		notification.AddLog(logx.GetLogPushEntry(&logx.InputLog{
			ID:        notification.ID,
			Status:    core.FailedPush,
			Token:     token,
			Message:   notification.Message,
			Platform:  notification.Platform,
			Error:     errors.New(reason),
			HideToken: cfg.Log.HideToken,
			Format:    cfg.Log.Format,
		}))
	}
	notification.WaitDone()
}

// HandleNotification add notification to queue list.
func handleNotification(ctx context.Context, cfg config.ConfYaml, req go_scrape.RequestPush, q *queue.Queue) (int, []logx.LogPushEntry) {
	var count int
	wg := sync.WaitGroup{}
	newNotification := []*go_scrape.PushNotification{}

	if cfg.Core.Sync && !core.IsLocalQueue(core.Queue(cfg.Queue.Engine)) {
		cfg.Core.Sync = false
	}

	for i := range req.Notifications {
		notification := &req.Notifications[i]
		notification.Cfg = cfg
		newNotification = append(newNotification, notification)
	}

	returnedLog := make([]logx.LogPushEntry, 0, count)
	for _, notification := range newNotification {
		if cfg.Core.Sync {
			notification.Wg = &wg
			notification.Log = &returnedLog
			notification.AddWaitCount()
		}

		if err := q.Queue(notification); err != nil {
			markFailedNotification(cfg, notification, "max capacity reached")
		}

		count += len(notification.Tokens)
		// Count topic message
		if notification.To != "" {
			count++
		}
	}

	if cfg.Core.Sync {
		wg.Wait()
	}

	status.StatStorage.AddTotalCount(int64(count))

	return count, returnedLog
}

func routerEngine(cfg config.ConfYaml, q *queue.Queue) *gin.Engine {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if cfg.Core.Mode == "debug" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

	if isTerm {
		log.Logger = log.Output(
			zerolog.ConsoleWriter{
				Out:     os.Stdout,
				NoColor: false,
			},
		)
	}

	// Support metrics
	doOnce.Do(func() {
		m := metric.NewMetrics(func() int {
			return q.Usage()
		})
		prometheus.MustRegister(m)
	})

	// set server mode
	gin.SetMode(cfg.Core.Mode)

	r := gin.New()

	// Global middleware
	r.Use(logger.SetLogger(
		logger.WithUTC(true),
		logger.WithSkipPath([]string{
			cfg.API.HealthURI,
			cfg.API.MetricURI,
		}),
	))
	r.Use(gin.Recovery())
	r.Use(VersionMiddleware())
	r.Use(StatMiddleware())

	r.GET(cfg.API.StatGoURI, api.GinHandler)
	r.GET(cfg.API.StatAppURI, appStatusHandler(q))
	r.GET(cfg.API.ConfigURI, configHandler())
	r.GET(cfg.API.SysStatURI, sysStatsHandler())
	r.POST(cfg.API.PushURI, pushHandler(q))
	r.GET(cfg.API.MetricURI, metricsHandler)
	r.GET(cfg.API.HealthURI, heartbeatHandler)
	r.HEAD(cfg.API.HealthURI, heartbeatHandler)

	r.GET("/debugPush", debugPushHandler)
	r.GET("/version", versionHandler)
	r.GET("/", rootHandler)

	r.GET("/scrape/funds", scrapeFundHandler())
	r.GET("/scrape/nav/:id", scrapeNavHandler())

	return r
}
