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

func NewScheduler(cs check.ChecksV2, resCh chan check.ExecResult) *Scheduler {
	s := gocron.NewScheduler(time.UTC)
	for _, c := range cs {
		log.Debugf("Adding check to scheduler: %s", c.Name)
		_, err := s.Every(c.Interval).Seconds().Do(c.Run, resCh)
		if err != nil {
			log.Errorf("Error scheduling check: %s", err.Error())
		}
	}
	return &Scheduler{s}
}
