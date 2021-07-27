package config

import (
	"github.com/natansdj/go_scrape/logx"
	"net"
	"net/http"
	"time"
)

var (
	ScrHttpT     *http.Transport
	ScrNetClient *http.Client
)

func InitClient(cfg ConfYaml) {
	defer logx.LogAccess.Info("Init http.Client && http.Transport")

	//Init Config
	appConf := cfg.Source

	scrCtxTimeout := appConf.CtxTimeout
	if scrCtxTimeout == 0 {
		scrCtxTimeout = 30
	}
	scrCtxKeepAlive := appConf.CtxKeepAlive
	if scrCtxKeepAlive == 0 {
		scrCtxKeepAlive = 100
	}
	scrMaxIdleConnsPerHost := appConf.MaxIdleConnsPerHost
	if scrMaxIdleConnsPerHost == 0 {
		scrMaxIdleConnsPerHost = 100
	}
	scrMaxIdleConns := appConf.MaxIdleConns
	if scrMaxIdleConns == 0 {
		scrMaxIdleConns = 100
	}
	scrIdleConnTimeout := appConf.IdleConnTimeout
	if scrIdleConnTimeout == 0 {
		scrIdleConnTimeout = 90
	}
	scrTLSHandshakeTimeout := appConf.TLSHandshakeTimeout
	if scrTLSHandshakeTimeout == 0 {
		scrTLSHandshakeTimeout = 10
	}
	scrExpectContinueTimeout := appConf.ExpectContinueTimeout
	if scrExpectContinueTimeout == 0 {
		scrExpectContinueTimeout = 3
	}
	scrHttpTimeout := appConf.HttpTimeout
	if scrHttpTimeout == 0 {
		scrHttpTimeout = 5
	}

	//Init Transport & HTTP
	ScrHttpT = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   time.Duration(scrCtxTimeout) * time.Second,
			KeepAlive: time.Duration(scrCtxKeepAlive) * time.Second,
		}).DialContext,
		MaxIdleConnsPerHost:   scrMaxIdleConnsPerHost,
		MaxIdleConns:          scrMaxIdleConns,
		IdleConnTimeout:       time.Duration(scrIdleConnTimeout) * time.Second,
		TLSHandshakeTimeout:   time.Duration(scrTLSHandshakeTimeout) * time.Second,
		ExpectContinueTimeout: time.Duration(scrExpectContinueTimeout) * time.Second,
	}

	ScrNetClient = &http.Client{
		Transport: ScrHttpT,
		Timeout:   time.Duration(scrHttpTimeout) * time.Second,
	}
}
