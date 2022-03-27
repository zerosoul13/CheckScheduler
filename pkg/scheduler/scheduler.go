package scheduler

import (
	"mon-agent/pkg/check"
	"time"

	"github.com/go-co-op/gocron"

	log "github.com/sirupsen/logrus"
)

// Scheduler is a wrapper around gocron.Scheduler
// to create a scheduler with a list of checks
type Scheduler struct {
	*gocron.Scheduler
}

func (s *Scheduler) register(checks check.Checks, resCh chan check.ExecResult) {
	for _, check := range checks {
		log.Debugf("Registering check %s to scheduler", check.Name)
		err := s.Every(check.Interval, check.Run, resCh)
		if err != nil {
			log.Errorf("Error scheduling check: %s", err.Error())
		}
	}
}

// Schedule schedules checks and calls for execution immediately after start
func (s *Scheduler) Schedule(checks check.Checks, resCh chan check.ExecResult) {
	s.register(checks, resCh)

	s.StartImmediately()
	s.SingletonModeAll()
	s.StartBlocking()
}

// Every schedules checks to be executed every interval. Interval is expressed in seconds
func (s *Scheduler) Every(seconds int, f interface{}, resCh chan check.ExecResult) error {
	_, err := s.Scheduler.Every(seconds).Seconds().Do(f, resCh)
	return err
}

func NewScheduler() *Scheduler {
	s := gocron.NewScheduler(time.UTC)
	return &Scheduler{s}
}
