package main

import (
	"github.com/sdelicata/caeche/server"
	log "github.com/sirupsen/logrus"
	"os"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func main() {
	server := server.NewServer()
	log.Fatalln(server.Start())
}
