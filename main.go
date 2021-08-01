package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/natansdj/go_scrape/config"
	"github.com/natansdj/go_scrape/core"
	"github.com/natansdj/go_scrape/logx"
	"github.com/natansdj/go_scrape/models"
	"github.com/natansdj/go_scrape/queue"
	"github.com/natansdj/go_scrape/queue/simple"
	"github.com/natansdj/go_scrape/router"
	"github.com/natansdj/go_scrape/status"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

func withContextFunc(ctx context.Context, f func()) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(c)

		select {
		case <-ctx.Done():
		case <-c:
			cancel()
			f()
		}
	}()

	return ctx
}

func main() {
	opts := config.ConfYaml{}

	var (
		ping        bool
		showVersion bool
		configFile  string
	)

	flag.BoolVar(&showVersion, "version", false, "Print version information.")
	flag.BoolVar(&showVersion, "V", false, "Print version information.")
	flag.StringVar(&configFile, "c", "", "Configuration file path.")
	flag.StringVar(&configFile, "config", "", "Configuration file path.")
	flag.StringVar(&opts.Core.Port, "port", "", "port number for go_scrape")
	flag.StringVar(&opts.Stat.Engine, "e", "", "store engine")
	flag.StringVar(&opts.Stat.Engine, "engine", "", "store engine")
	flag.StringVar(&opts.Stat.Redis.Addr, "redis-addr", "", "redis addr")
	flag.BoolVar(&ping, "ping", false, "ping server")

	flag.Usage = usage
	flag.Parse()

	router.SetVersion(Version)

	// Show version and exit
	if showVersion {
		router.PrintVersion()
		os.Exit(0)
	}

	// set default parameters.
	cfg, err := config.LoadConf(configFile)
	if err != nil {
		log.Printf("Load yaml config file error: '%v'", err)
		return
	}

	if err = logx.InitLog(
		cfg.Log.AccessLevel,
		cfg.Log.AccessLog,
		cfg.Log.ErrorLevel,
		cfg.Log.ErrorLog,
	); err != nil {
		log.Fatalf("Can't load log module, error: %v", err)
	}

	if ping {
		if err := pinger(cfg); err != nil {
			logx.LogError.Warnf("ping server error: %v", err)
		}
		return
	}

	if opts.Core.PID.Path != "" {
		cfg.Core.PID.Path = opts.Core.PID.Path
		cfg.Core.PID.Enabled = true
		cfg.Core.PID.Override = true
	}

	if err = createPIDFile(cfg); err != nil {
		logx.LogError.Fatal(err)
	}

	if err = status.InitAppStatus(cfg); err != nil {
		logx.LogError.Fatal(err)
	}

	// Initialize Client
	config.InitClient(cfg)

	// Initialize DB
	models.ConnectDatabase()

	var w queue.Worker
	switch core.Queue(cfg.Queue.Engine) {
	case core.LocalQueue:
		w = simple.NewWorker(
			simple.WithQueueNum(int(cfg.Core.QueueNum)),
		)
	//case core.NSQ:
	//	w = nsq.NewWorker()
	default:
		logx.LogError.Fatalf("we don't support queue engine: %s", cfg.Queue.Engine)
	}

	q := queue.NewQueue(w, int(cfg.Core.WorkerNum))
	q.Start()

	finished := make(chan struct{})
	ctx := withContextFunc(context.Background(), func() {
		logx.LogAccess.Info("close the queue system")
		// stop queue system
		//q.Shutdown()
		// wait job completed
		//q.Wait()
		close(finished)
		// close the connection with storage
		logx.LogAccess.Infof("close the storage connection: %v", cfg.Stat.Engine)
		if err := status.StatStorage.Close(); err != nil {
			logx.LogError.Fatalf("can't close the storage connection: %v", err.Error())
		}
	})

	var g errgroup.Group
	// Run httpd server
	g.Go(func() error {
		return router.RunHTTPServer(ctx, cfg, q)
	})

	// check job completely
	g.Go(func() error {
		select {
		case <-finished:
		}
		return nil
	})

	if err = g.Wait(); err != nil {
		logx.LogError.Fatal(err)
	}
}

// Version control.
var Version = "1.0.0"

var usageStr = `
  ________                              .__
 /  _____/   ____  
/   \  ___  /  _ \ 
\    \_\  \(  <_> )
 \______  / \____/ 
        \/         

Usage: go_scrape [options]

Server Options:
    -A, --address <address>          Address to bind (default: any)
    -p, --port <port>                Use port for clients (default: 8088)
    -c, --config <file>              Configuration file path
    -m, --message <message>          Notification message
    -t, --token <token>              Notification token
    -e, --engine <engine>            Storage engine (memory, redis ...)
    --title <title>                  Notification title
    --proxy <proxy>                  Proxy URL
    --pid <pid path>                 Process identifier path
    --redis-addr <redis addr>        Redis addr (default: localhost:6379)
    --ping                           healthy check command for container
    -h, --help                       Show this message
    -V, --version                    Show version
`

// usage will print out the flag options for the server.
func usage() {
	fmt.Printf("%s\n", usageStr)
	os.Exit(0)
}

// handles pinging the endpoint and returns an error if the
// agent is in an unhealthy state.
func pinger(cfg config.ConfYaml) error {
	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	client := &http.Client{
		Timeout:   time.Second * 10,
		Transport: transport,
	}
	resp, err := client.Get("http://localhost:" + cfg.Core.Port + cfg.API.HealthURI)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned non-200 status code")
	}
	return nil
}

func createPIDFile(cfg config.ConfYaml) error {
	if !cfg.Core.PID.Enabled {
		return nil
	}

	pidPath := cfg.Core.PID.Path
	_, err := os.Stat(pidPath)
	if os.IsNotExist(err) || cfg.Core.PID.Override {
		currentPid := os.Getpid()
		if err := os.MkdirAll(filepath.Dir(pidPath), os.ModePerm); err != nil {
			return fmt.Errorf("Can't create PID folder on %v", err)
		}

		file, err := os.Create(pidPath)
		if err != nil {
			return fmt.Errorf("Can't create PID file: %v", err)
		}
		defer file.Close()
		if _, err := file.WriteString(strconv.FormatInt(int64(currentPid), 10)); err != nil {
			return fmt.Errorf("Can't write PID information on %s: %v", pidPath, err)
		}
	} else {
		return fmt.Errorf("%s already exists", pidPath)
	}
	return nil
}
