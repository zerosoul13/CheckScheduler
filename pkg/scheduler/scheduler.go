package scheduler

import (
	"time"

	"github.com/go-co-op/gocron"
)

type Scheduler struct {
	*gocron.Scheduler
}

func NewScheduler() *gocron.Scheduler {
	return gocron.NewScheduler(time.UTC)
}
