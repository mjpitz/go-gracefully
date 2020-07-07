package report

import (
	"github.com/mjpitz/go-gracefully/check"
)

// CheckResult is a static capture of a check and associated results.
type CheckResult struct {
	check.Metadata
	LastCheck      check.Result `json:"last_check"`
	LastKnownCheck check.Result `json:"last_known_check"`
}

// Report is a static capture of an application and associated results.
type Report struct {
	check.Result
	Results map[string]CheckResult `json:"results"`
}
