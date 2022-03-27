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

const CHECKSFILE = "checks.json"

type Command struct {
	name    string
	command string
	args    []string
	hasArgs bool
}

func (c *Command) Exec() ExecResult {
	var out []byte
	var err error

	start := time.Now()
	if !c.hasArgs {
		out, err = exec.Command(c.command).CombinedOutput()
	} else {
		out, err = exec.Command(c.command, c.args...).CombinedOutput()
	}

	return ExecResult{
		Name:     c.name,
		Error:    err,
		Stdout:   string(out),
		ExecTime: time.Since(start).Seconds(),
		PerfData: string(out),
	}

}

func NewCommand(name string, cmd string, args []string) *Command {
	return &Command{
		name:    name,
		command: cmd,
		args:    args,
		hasArgs: len(args) > 0,
	}
}

func NewCommandFromString(name string, cmd string) *Command {
	s := strings.Split(cmd, " ")

	if len(s) >= 2 {
		return NewCommand(name, s[0], s[1:])
	}

	var args []string
	return NewCommand(name, s[0], args)
}

// Checks holds a map of checks
type Checks map[string]Check

func (c Checks) String() string {
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

	// Command is the command to be executed in string format
	Command string

	// Interval is the interval in seconds to execute the check
	Interval int

	// Timeout is the timeout in seconds to wait for the check to complete
	Timeout int64

	// cmd is the command to be executed by the check
	cmd *Command
}

// Run executes the check defined command
func (c Check) Run(resCh chan ExecResult) {
	// done	is a channel to signal the end of the check
	done := make(chan bool)

	if c.cmd == nil {
		c.cmd = NewCommandFromString(c.Name, c.Command)
	}

	var res ExecResult
	start := time.Now()
	go func() {
		log.Debugf("Calling check: %s", c.Name)

		res = c.cmd.Exec()
		log.Debugf("Results for check: %s have been submitted. %v", c.Name, res)
		resCh <- res
		done <- true
	}()

	select {
	case <-time.After(time.Duration(c.Timeout) * time.Second):
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

// GetChecks returns a map of checks
func GetChecks() (Checks, error) {
	var err error
	checks := make(Checks)

	// read json file and unmarshal to Check type
	file, err := os.Open(CHECKSFILE)
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
