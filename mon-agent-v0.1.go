package main

import (
	"mon-agent/pkg/check"
	"mon-agent/pkg/scheduler"

	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)
	log.Info("Starting mon-agent")
	// resCh is a channel to receive the results of the checks
	resCh := make(chan check.ExecResult)

	// Read the results of the checks
	go func() {
		check.Read(resCh)
	}()

	// Collect checks based on host identification
	checks, err := check.GetChecks(resCh)
	if err != nil {
		log.Fatal("Error loading Checks: ", err)
	} else {
		log.Info("Checks to execute: ", checks.String())
	}

	s := scheduler.NewScheduler()

	var jobs []*gocron.Job
	for _, c := range checks {
		job, err := s.Every(c.Interval).Seconds().Do(c.Run)
		if err != nil {
			log.Errorf("Error scheduling check: %s", err.Error())
		}
		jobs = append(jobs, job)
	}

	go func(jobs []*gocron.Job) {
		for _, job := range jobs {
			log.Debugf("Starting job: %s", job.NextRun().Format("2006-01-02 15:04:05"))
		}
	}(jobs)

	s.StartImmediately()
	s.SingletonModeAll()
	s.StartBlocking()

	log.Info("All checks have completed")

	// Close the channel
	close(resCh)
}
