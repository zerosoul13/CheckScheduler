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
		log.Warnf("No perfdata returned by check: %s", r.Name)
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

	ps, _ := exec.LookPath("cmd")

	cmd, err := exec.Command(ps + " " + c.Command).CombinedOutput()
	res.Name = c.Name
	res.Error = err
	res.Stdout = string(cmd)
	res.ExecTime = time.Since(start).Seconds()
	log.Debugf("Results for check: %s have been sent through chan", c.Name)

	c.Result <- res

	// We must cancel the timeout goroutine otherwise it will produce odd timeouts
	//
	// {"level":"info","msg":"Starting mon-agent","time":"2022-03-24T21:22:56-07:00"}
	// {"level":"info","msg":"Checks to execute: disk, xyz, xfz, xaz, xzz, xsz, disk, ","time":"2022-03-24T21:22:56-07:00"}
	// {"level":"info","msg":"Error executing check: exec: \"C:/Users/angelf.rodriguez/Desktop/golang/mock/check_timeout.bat' \": file does not exist","time":"2022-03-24T21:22:56-07:00"}
	// {"level":"info","msg":"Error executing check: timeout executing check: disk","time":"2022-03-24T21:23:06-07:00"}

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
