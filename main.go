package main

import (
	"crypto/tls"
	"github.com/justinas/alice"
	"github.com/sdelicata/caeche/cache"
	"github.com/sdelicata/caeche/config"
	"github.com/sdelicata/caeche/server"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"net/http"
	"os"
	"time"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func main() {
	cfg, err := config.NewConfigFromFile("config.toml")
	if err != nil {
		log.Errorf("Error loading config file : %s", err)
		return
	}

	initTransport(cfg.Backend.Scheme)

	cacheInMemory := cache.NewInMemory(cfg.DefaultTTL)
	reverseProxy := server.NewReverseProxy(cfg, cacheInMemory)
	purgeMiddleWare := cache.NewPurgeMiddleware(cacheInMemory)

	chain := alice.New(purgeMiddleWare).Then(reverseProxy.GetHandler())

	s := http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      chain,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}
	log.Infoln("Server starting...")
	if cfg.Backend.Scheme == "https" {
		log.Fatalln(s.ListenAndServeTLS("cert.pem", "key.pem"))
	} else {
		log.Fatalln(s.ListenAndServe())
	}
}

func initTransport(scheme string) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	if scheme == "https" {
		err := http2.ConfigureTransport(http.DefaultTransport.(*http.Transport))
		if err != nil {
			log.Error(err)
		}
	}
}
