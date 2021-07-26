package config

import (
	"github.com/natansdj/go_scrape/logx"
	"net"
	"net/http"
	"time"
)

var (
	DdcHttpT     *http.Transport
	DdcNetClient *http.Client
)

func InitClient(cfg ConfYaml) {
	logx.LogAccess.Info("Init http.Client && http.Transport")

	//Init Config
	appConf := cfg.Source

	ddcCtxTimeout := appConf.CtxTimeout
	if ddcCtxTimeout == 0 {
		ddcCtxTimeout = 30
	}
	ddcCtxKeepAlive := appConf.CtxKeepAlive
	if ddcCtxKeepAlive == 0 {
		ddcCtxKeepAlive = 100
	}
	ddcMaxIdleConnsPerHost := appConf.MaxIdleConnsPerHost
	if ddcMaxIdleConnsPerHost == 0 {
		ddcMaxIdleConnsPerHost = 100
	}
	ddcMaxIdleConns := appConf.MaxIdleConns
	if ddcMaxIdleConns == 0 {
		ddcMaxIdleConns = 100
	}
	ddcIdleConnTimeout := appConf.IdleConnTimeout
	if ddcIdleConnTimeout == 0 {
		ddcIdleConnTimeout = 90
	}
	ddcTLSHandshakeTimeout := appConf.TLSHandshakeTimeout
	if ddcTLSHandshakeTimeout == 0 {
		ddcTLSHandshakeTimeout = 10
	}
	ddcExpectContinueTimeout := appConf.ExpectContinueTimeout
	if ddcExpectContinueTimeout == 0 {
		ddcExpectContinueTimeout = 3
	}

	ddcHttpTimeout := appConf.HttpTimeout
	if ddcHttpTimeout == 0 {
		ddcHttpTimeout = 5
	}

	//Init Transport & HTTP
	DdcHttpT = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   time.Duration(ddcCtxTimeout) * time.Second,
			KeepAlive: time.Duration(ddcCtxKeepAlive) * time.Second,
		}).DialContext,
		MaxIdleConnsPerHost:   ddcMaxIdleConnsPerHost,
		MaxIdleConns:          ddcMaxIdleConns,
		IdleConnTimeout:       time.Duration(ddcIdleConnTimeout) * time.Second,
		TLSHandshakeTimeout:   time.Duration(ddcTLSHandshakeTimeout) * time.Second,
		ExpectContinueTimeout: time.Duration(ddcExpectContinueTimeout) * time.Second,
	}

	DdcNetClient = &http.Client{
		Transport: DdcHttpT,
		Timeout:   time.Duration(ddcHttpTimeout) * time.Second,
	}
}
