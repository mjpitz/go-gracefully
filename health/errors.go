package health

import "fmt"

var (
	// ErrAlreadyStarted is returned when the Monitor has alreadybeen started
	ErrAlreadyStarted = fmt.Errorf("monitor already started")
)
