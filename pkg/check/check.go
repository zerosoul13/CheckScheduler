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

type ChecksV2 map[string]Check

func (c ChecksV2) String() string {
	s := ""

	for _, checkName := range c {
		s += checkName.Name + ", "
	}

	return s
}

// Check is a wrapper for a command
// used to monitor a system
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
}

// Run executes the check defined command
func (c Check) Run(resCh chan ExecResult) {
	// done	is a channel to signal the end of the check
	done := make(chan bool)

	var res ExecResult
	start := time.Now()
	go func() {
		log.Debugf("Calling check: %s", c.Name)

		var cmd []byte
		var err error

		cm, args, hasArgs := c.prepare()
		log.Debugf("Command: %s, Args: %s", cm, args)
		if !hasArgs {
			// This command does not have any extra arguments
			cmd, err = exec.Command(cm).CombinedOutput()
		} else {
			cmd, err = exec.Command(cm, args...).CombinedOutput()
		}

		res = ExecResult{
			Name:     c.Name,
			Error:    err,
			Stdout:   string(cmd),
			ExecTime: time.Since(start).Seconds(),
			PerfData: string(cmd),
		}

		log.Debugf("Results for check: %s have been submitted. %v", c.Name, res)
		resCh <- res
		done <- true
	}()

	timeout := time.After(time.Duration(c.Timeout) * time.Second)
	select {
	case <-timeout:
		e := fmt.Errorf("Timeout for check: %s", c.Name)
		res.Error = e
		res.ExecTime = time.Since(start).Seconds()
		res.Stdout = e.Error()
		res.PerfData = e.Error()
		res.Name = c.Name
		resCh <- res

	case <-done:
		log.Infof("%s has completed in %f seconds", c.Name, res.ExecTime)
	}
}

// prepare prepares the command to be executed
// returns the command and arguments
func (c Check) prepare() (string, []string, bool) {
	s := strings.Split(c.Command, " ")

	if len(s) >= 2 {
		return s[0], s[1:], true
	}

	var arg []string
	return s[0], arg, false
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
func GetChecks() (ChecksV2, error) {
	var err error
	var checks ChecksV2

	checks = make(ChecksV2)

	// read json file and unmarshal to Check type
	f := "checksv2.json"
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
	}

	return checks, err
}
