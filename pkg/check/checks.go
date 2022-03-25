package check

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// ExecResult is the result of a check
type ExecResult struct {

	// Name is the name of the check
	Name string

	// Stdout is the standard output of the check
	Stdout string

	// Error is the error returned by the check
	Error error

	// ExecTime is the time it took to execute the check
	ExecTime float64

	// PerfData is the performance data returned by the check
	PerfData string
}

func (r ExecResult) Result() (string, string) {

	// no perfdata
	if !strings.Contains(r.Stdout, "|") {
		return r.Stdout, ""
	}

	split := strings.Split(r.Stdout, "|")

	// split[0] is the output of the check
	r.Stdout = split[0]

	// split[1] is the perfdata in raw format
	r.PerfData = split[1]

	return r.Stdout, perfdata(r.PerfData)
}

// Check is a single check to be executed
// to monitor a system
type Check struct {
	// Name is the name of the check
	Name string

	// Description is the description of the check
	Description string

	// Command is the command to be executed
	Command string

	// Interval is the interval in seconds to execute the check
	Interval int

	// Timeout is the timeout in seconds to wait for the check to complete
	Timeout int64

	// Result is the channel to send the result of the check
	Result chan ExecResult
}

// Run executes the check defined command
func (c Check) Run() {
	res := ExecResult{}

	start := time.Now()
	log.Debugf("Calling check: %s", c.Name)

	// Protect execution from running for too long.
	// Each command has a timeout option that can be adjusted in the config file.

	cancelCh := make(chan struct{})

	go func(cancel chan struct{}) {
		timeout := time.After(time.Duration(c.Timeout) * time.Second)

		select {
		case <-timeout:
			log.Debugf("Timeout executing check: %s", c.Name)
			res.Error = fmt.Errorf("timeout executing check: %s", c.Name)
			res.Name = c.Name
			c.Result <- res
		case <-cancel:
			log.Debugf("Cancelled executing check: %s", c.Name)
			return
		}
	}(cancelCh)

	cm := strings.Split(c.Command, " ")

	var cmd []byte
	var err error
	if len(cm) < 2 {
		cmd, err = exec.Command(cm[0]).CombinedOutput()
	} else {
		cmd, err = exec.Command(cm[0], cm[1:]...).CombinedOutput()
	}

	res.Name = c.Name
	res.Error = err
	res.Stdout = string(cmd)
	res.ExecTime = time.Since(start).Seconds()
	log.Debugf("Results for check: %s have been sent through chan", c.Name)

	c.Result <- res

	log.Debugf("Cancelling check timeout timer: %s", c.Name)
	cancelCh <- struct{}{}
}

// Checks is a slice of checks
type Checks []Check

// String returns a string representation of the checks
func (c Checks) String() string {
	s := ""

	for _, checkName := range c {
		s += checkName.Name + ", "
	}

	return s
}

// GetChecks returns a slice of checks
func GetChecks(resCh chan ExecResult) (Checks, error) {
	var err error
	var checks Checks

	// read json file and unmarshal to Check type
	f := "checks.json"
	file, err := os.Open(f)
	if err != nil {
		log.Errorf("Error opening file: %s", err)
		return checks, err
	} else {
		err = json.NewDecoder(file).Decode(&checks)
		if err != nil {
			log.Errorf("Error decoding file: %s", err)
			return checks, err
		}

		// set the result channel for each check
		for i, _ := range checks {
			checks[i].Result = resCh
		}

	}
	return checks, err
}
