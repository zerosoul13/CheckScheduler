package manager

import (
	"mon-agent/pkg/check"
	"mon-agent/pkg/output"
	"mon-agent/pkg/scheduler"
	"mon-agent/pkg/tsdb"

	log "github.com/sirupsen/logrus"
)

type IScheduler interface {
	Schedule(checks check.Checks, resCh chan check.ExecResult)
	Every(seconds int, f interface{}, resCh chan check.ExecResult) error
}

type IPublisher interface {
	Publish(resCh chan check.ExecResult)
}

func NewManager(c *tsdb.Graphite) *Manager {
	resCh := make(chan check.ExecResult)

	// Collect checks based on host identification
	checks, err := check.GetChecks()
	if err != nil {
		log.Fatal("Error loading Checks: ", err)
	} else {
		log.Debugf("Checks to execute: %s", checks.String())
	}

	return &Manager{
		scheduler: scheduler.NewScheduler(),
		publisher: output.NewPublisher(c),
		checks:    checks,
		resCh:     resCh,
	}
}

// Manager is a wrapper around gocron.Scheduler
// to create a scheduler with a list of checks
// and execute them.
type Manager struct {
	// The scheduler
	scheduler IScheduler

	// publisher is used to publish results
	// to Graphite
	publisher IPublisher

	// The channel to receive the results of the checks
	resCh chan check.ExecResult

	// The list of checks to be executed
	checks check.Checks
}

func (m *Manager) Start() {
	// Publish results to Graphite
	go func() {
		m.publisher.Publish(m.resCh)
	}()

	// Display results from the channel to user
	go func() {
		m.Display()
	}()

	// start the scheduler
	m.scheduler.Schedule(m.checks, m.resCh)
}

func (m *Manager) Stop() {
	close(m.resCh)
}

// Display prints the results received from the channel
func (m *Manager) Display() {
	for message := range m.resCh {
		output, pdata := message.Result()
		if message.Error != nil {
			log.WithFields(log.Fields{
				"check":    message.Name,
				"exectime": message.ExecTime,
				"output":   output,
				"error":    message.Error,
			}).Error(output)
		} else {
			log.WithFields(log.Fields{
				"check":    message.Name,
				"exectime": message.ExecTime,
				"output":   output,
				"perfdata": pdata,
			}).Info(output)
		}
	}
}
