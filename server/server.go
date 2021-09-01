package server

import (
	"github.com/sdelicata/caeche/config"
	"net/http"
	"time"
)

type Server http.Server

func NewServer(config config.Config) *http.Server {
	reverseProxy := NewReverseProxy(config)

	return &http.Server{
		Addr: ":" + config.Port,
		Handler: reverseProxy,
		WriteTimeout: 30 * time.Second,
		ReadTimeout: 30 * time.Second,
	}
}


