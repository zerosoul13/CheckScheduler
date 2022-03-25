package check

// ExecuteChecks executes a slice of checks
func ExecuteChecks(checks Checks, resCh chan ExecResult) {
	for _, check := range checks {
		go check.Run()
	}
}
