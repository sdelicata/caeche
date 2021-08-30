package server

import (
	"crypto/tls"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type Server struct {
	proxy *httputil.ReverseProxy
	httpServer *http.Server
}

func (server *Server) Start() error {
	log.Debugln("Server starting...")
	return server.httpServer.ListenAndServe()
}

func NewServer() *Server {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("google.com"),
	})

	httpServer := &http.Server{
		Addr:         ":8080",
		Handler:      proxy,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	server := &Server{
		proxy: proxy,
		httpServer: httpServer,
	}

	return server
}


