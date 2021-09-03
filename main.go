package main

import (
	"github.com/justinas/alice"
	"github.com/sdelicata/caeche/cache"
	"github.com/sdelicata/caeche/config"
	"github.com/sdelicata/caeche/server"
	log "github.com/sirupsen/logrus"
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

	cacheInMemory := cache.NewInMemory(cfg.Cache.DefaultTTL)
	reverseProxy := server.NewReverseProxy(cfg, cacheInMemory)
	purgeMiddleWare := cache.NewPurgeMiddleware(cacheInMemory)

	chain := alice.New(purgeMiddleWare).Then(reverseProxy)

	s := http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      chain,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}
	log.Infoln("Server starting...")
	log.Fatalln(s.ListenAndServe())
}
