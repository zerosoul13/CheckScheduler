package check

import "strings"

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

	// no perfdata found
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
