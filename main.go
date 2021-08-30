package main

import (
	"github.com/sdelicata/caeche/config"
	"github.com/sdelicata/caeche/server"
	log "github.com/sirupsen/logrus"
	"os"
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

	server := server.NewServer(config)
	log.Fatalln(server.Start())
}
