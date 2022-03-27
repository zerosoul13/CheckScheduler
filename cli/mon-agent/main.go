package main

import (
	"flag"
	"mon-agent/pkg/manager"
	"mon-agent/pkg/tsdb"
	"time"

	log "github.com/sirupsen/logrus"
)

// loadProfile loads the profile from the file
func loadProfile() {
	log.Info("Loading profile..")
}

func main() {
	level := flag.Bool("debug", false, "Enable debug mode")
	gHost := flag.String("graphite-host", "localhost", "Graphite host")
	gPort := flag.String("graphite-port", "2003", "Graphite port")

	flag.Parse()

	if *level {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	log.Info("Starting mon-agent..")

	// TODO: Implement
	loadProfile()

	// Results collected are sent to the Graphite through publisher
	c := tsdb.NewGraphite(*gHost, *gPort, "tcp", time.Duration(10*time.Second))

	// Create a new manager with the list of checks and the publisher
	manager := manager.NewManager(c)
	manager.Start()
}
