package check

import "fmt"

var (
	// ErrTimeout is returned when a check times out during evaluation
	ErrTimeout = fmt.Errorf("timed out waiting for check")
)
