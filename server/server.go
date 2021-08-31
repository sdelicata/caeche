package server

import (
	"github.com/justinas/alice"
	"github.com/sdelicata/caeche/config"
	"net/http"
	"time"
)

type Server http.Server

func NewServer(config config.Config) *http.Server {
	reverseProxy := NewReverseProxy(config)
	cacheMiddleware := NewCacheMiddleware()
	chain := alice.New(cacheMiddleware).Then(reverseProxy)

	return &http.Server{
		Addr: ":" + config.Port,
		Handler: chain,
		WriteTimeout: 30 * time.Second,
		ReadTimeout: 30 * time.Second,
	}
}


