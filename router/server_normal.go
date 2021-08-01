// +build !lambda

package router

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"github.com/natansdj/go_scrape/queue"
	"net/http"
	"time"

	"github.com/natansdj/go_scrape/config"
	"github.com/natansdj/go_scrape/logx"

	"golang.org/x/sync/errgroup"
)

// RunHTTPServer provide run http or https protocol.
func RunHTTPServer(ctx context.Context, cfg config.ConfYaml, q *queue.Queue, s ...*http.Server) (err error) {
	var server *http.Server

	if !cfg.Core.Enabled {
		logx.LogAccess.Info("httpd server is disabled.")
		return nil
	}

	if len(s) == 0 {
		server = &http.Server{
			Addr:    cfg.Core.Address + ":" + cfg.Core.Port,
			Handler: routerEngine(cfg, q),
		}
	} else {
		server = s[0]
	}

	defer logx.LogAccess.Infof("HTTPD server is running on %v port.", cfg.Core.Port)
	if cfg.Core.AutoTLS.Enabled {
		return startServer(ctx, autoTLSServer(cfg, q), cfg)
	} else if cfg.Core.SSL {
		conf := &tls.Config{
			MinVersion: tls.VersionTLS10,
		}

		if conf.NextProtos == nil {
			conf.NextProtos = []string{"http/1.1"}
		}

		conf.Certificates = make([]tls.Certificate, 1)
		if cfg.Core.CertPath != "" && cfg.Core.KeyPath != "" {
			conf.Certificates[0], err = tls.LoadX509KeyPair(cfg.Core.CertPath, cfg.Core.KeyPath)
			if err != nil {
				logx.LogError.Error("Failed to load https cert file: ", err)
				return err
			}
		} else if cfg.Core.CertBase64 != "" && cfg.Core.KeyBase64 != "" {
			cert, err := base64.StdEncoding.DecodeString(cfg.Core.CertBase64)
			if err != nil {
				logx.LogError.Error("base64 decode error:", err.Error())
				return err
			}
			key, err := base64.StdEncoding.DecodeString(cfg.Core.KeyBase64)
			if err != nil {
				logx.LogError.Error("base64 decode error:", err.Error())
				return err
			}
			if conf.Certificates[0], err = tls.X509KeyPair(cert, key); err != nil {
				logx.LogError.Error("tls key pair error:", err.Error())
				return err
			}
		} else {
			return errors.New("missing https cert config")
		}

		server.TLSConfig = conf
	}

	return startServer(ctx, server, cfg)
}

func listenAndServe(ctx context.Context, s *http.Server, cfg config.ConfYaml) error {
	var g errgroup.Group
	g.Go(func() error {
		select {
		case <-ctx.Done():
			timeout := time.Duration(cfg.Core.ShutdownTimeout) * time.Second
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			return s.Shutdown(ctx)
		}
	})
	g.Go(func() error {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})
	return g.Wait()
}

func listenAndServeTLS(ctx context.Context, s *http.Server, cfg config.ConfYaml) error {
	var g errgroup.Group
	g.Go(func() error {
		select {
		case <-ctx.Done():
			timeout := time.Duration(cfg.Core.ShutdownTimeout) * time.Second
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			return s.Shutdown(ctx)
		}
	})
	g.Go(func() error {
		if err := s.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})
	return g.Wait()
}

func startServer(ctx context.Context, s *http.Server, cfg config.ConfYaml) error {
	if s.TLSConfig == nil {
		return listenAndServe(ctx, s, cfg)
	}

	return listenAndServeTLS(ctx, s, cfg)
}
