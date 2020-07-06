package state

// State refers the the current state of a given check.
type State string

const (
	// Unknown state is when a check cannot be evaluated.
	Unknown State = "unknown"
	// Outage state is when a check fails and is unhealthy.
	Outage State = "outage"
	// Major state is when a check fails, but it has time before it's critical issue.
	Major State = "major"
	// Minor state is when a check fails, but does not impact the over all behavior of the system.
	Minor State = "minor"
	// OK state is when the check is operating as expected.
	OK State = "ok"
)

// Score computes an adjustment factor based on the provided state. This
// factor is used when computing the overall health of a given system.
func Score(state State) float32 {
	switch state {
	case Unknown:
		return 0
	case Outage:
		return 0.25
	case Major:
		return 0.50
	case Minor:
		return 0.75
	case OK:
		return 1.00
	default:
		return 0
	}
}

// ForScore takes a given score and maps it back to the state. A valid score
// is any value between [0, 1] (inclusively on the ends). This function will
// never return an `Unknown` state.
func ForScore(score float32) State {
	if score <= 0.25 {
		return Outage
	} else if score <= 0.50 {
		return Major
	} else if score <= 0.75 {
		return Minor
	}
	return OK
}
