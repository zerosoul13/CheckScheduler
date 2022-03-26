package main

import (
	"flag"
	"mon-agent/pkg/check"
	"mon-agent/pkg/scheduler"

	log "github.com/sirupsen/logrus"
)

func main() {
	level := flag.String("loglevel", "info", "Log level (info or debug)")
	flag.Parse()

	if *level == "info" || *level != "debug" {
		log.SetLevel(log.InfoLevel)
	}

	if *level == "debug" {
		log.SetLevel(log.DebugLevel)
	}

	log.Info("Starting mon-agent")

	// resCh is a channel to receive the results of the checks
	// We read the results of the checks from this channel
	resCh := make(chan check.ExecResult)
	go func() {
		check.Publish(resCh)

	}()

	go func() {
		check.Read(resCh)
	}()

	// Collect checks based on host identification
	checks, err := check.GetChecks()
	if err != nil {
		log.Fatal("Error loading Checks: ", err)
	} else {
		log.Debugf("Checks to execute: %s", checks.String())
	}

	// Schedule the checks and collect the jobs
	s := scheduler.NewScheduler(checks, resCh)

	s.StartImmediately()
	s.SingletonModeAll()
	s.StartBlocking()

	// Close the channel
	close(resCh)
}
