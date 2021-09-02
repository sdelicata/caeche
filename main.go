package main

import (
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
	config, err := config.NewConfigFromFile("config.toml")
	if err != nil {
		log.Errorf("Error loading config file : %s", err)
		return
	}

	cacheInMemory := cache.NewInMemory(config.Cache.DefaultTTL)
	reverseProxy := server.NewReverseProxy(config, cacheInMemory)

	server := http.Server{
		Addr: ":" + config.Port,
		Handler: reverseProxy,
		WriteTimeout: 30 * time.Second,
		ReadTimeout: 30 * time.Second,
	}
	log.Infoln("Server starting...")
	log.Fatalln(server.ListenAndServe())
}
