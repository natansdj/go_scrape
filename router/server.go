package router

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/natansdj/go_scrape/metric"
	"github.com/natansdj/go_scrape/status"
	"net/http"
	"net/url"
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
		Status:    core.FailedPush,
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

func pushHandler(cfg config.ConfYaml, q *queue.Queue) gin.HandlerFunc {
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

func configHandler(cfg config.ConfYaml) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.YAML(http.StatusCreated, cfg)
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
	r.GET(cfg.API.ConfigURI, configHandler(cfg))
	r.GET(cfg.API.SysStatURI, sysStatsHandler())
	r.POST(cfg.API.PushURI, pushHandler(cfg, q))
	r.GET(cfg.API.MetricURI, metricsHandler)
	r.GET(cfg.API.HealthURI, heartbeatHandler)
	r.HEAD(cfg.API.HealthURI, heartbeatHandler)

	r.GET("/debugPush", debugPushHandler)
	r.GET("/version", versionHandler)
	r.GET("/", rootHandler)

	r.GET("/scrape/1", scrapeOneHandler(cfg))

	return r
}

func scrapeOneHandler(cfg config.ConfYaml) gin.HandlerFunc {
	return func(c *gin.Context) {

		baseUri := cfg.Source.BaseURI

		form := url.Values{}
		form.Add("firstopen", "yes")
		form.Add("aumlowervalue", "500")
		form.Add("aumlowercheck", "yes")
		form.Add("aumbetweenlowvalue", "500")
		form.Add("aumbetweenhighvalue", "2000")
		form.Add("aumbetweencheck", "yes")
		form.Add("aumgreatervalue", "2000")
		form.Add("aumgreatercheck", "yes")
		form.Add("availibility", "available")
		form.Add("fundtype", "mm,fi,balance,equity")
		form.Add("hiloselect", "1yr")
		form.Add("performancetype", "nav")
		form.Add("fundnonsyariah", "yes")
		form.Add("fundsyariah", "yes")
		form.Add("etfnonsyariah", "yes")
		form.Add("etfsyariah", "yes")

		req, _ := RequestInit(cfg, "GET", "source_json_for_favorite.php", nil, form)
		body, err := RequestDo(req)
		if err != nil {
			logx.LogError.Error(err.Error())
			panic(err)
		}

		j := NewJSONReader(body)
		var i gin.H
		dec := json.NewDecoder(j)
		err = dec.Decode(&i)
		if err != nil {
			logx.LogError.Error(err.Error())
			panic(err.Error())
		} else {
			fmt.Println(fmt.Sprintf("\nType : %T", i["aaData"]))

			//list of funds
			if aaData, ok := i["aaData"].([]interface{}); ok {
				fmt.Println(fmt.Sprintf("Len : %v", len(aaData)))
				for k, aDtRaw := range aaData {
					fmt.Println(fmt.Sprintf("%v, %T", aaData[k], aaData[k]))

					//Process each fund
					if aDtVal, ok2 := aDtRaw.([]interface{}); ok2 {
						for l := range aDtVal {
							switch l {
							case 0: //id
							case 1: //id
							case 2: //name
							case 3: //manager
							case 4: //type
							case 5: //last_nav
							case 6: //1d
							case 7: //3d
							case 8: //1m
							case 9: //3m
							case 10: //6m
							case 11: //9m
							case 12: //ytd
							case 13: //1yr
							case 14: //3yr
							case 15: //5yr
							case 16: //hi-lo
							case 17: //sharpe
							case 18: //drawdown
							case 19: //dd_periode
							case 20: //
							case 21: //
							case 22: //hist_risk
							case 23: //aum
							case 24: //
							case 25: //
							case 26: //
							case 27: //
							case 28: //
							case 29: //
							case 30: //
							case 31: //
							}
							fmt.Println(fmt.Sprintf("%v, %T", aDtVal[l], aDtVal[l]))
						}
					}
					break
				}
			}
			fmt.Println("")
		}

		c.JSON(http.StatusOK, gin.H{
			"source":  "https://www.indopremier.com/programer_script/source_json_for_favorite.php?firstopen=yes&aumlowervalue=500&aumlowercheck=yes&aumbetweenlowvalue=500&aumbetweenhighvalue=2000&aumbetweencheck=yes&aumgreatervalue=2000&aumgreatercheck=yes&availibility=available&fundtype=mm,fi,balance,equity,&hiloselect=1yr&performancetype=nav&fundnonsyariah=yes&fundsyariah=yes&etfnonsyariah=yes&etfsyariah=yes",
			"version": GetVersion(),
			"baseUri": baseUri,
			"result":  i,
		})
	}
}
