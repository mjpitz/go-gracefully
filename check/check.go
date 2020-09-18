package check

import (
	"context"
	"time"

	"github.com/mjpitz/go-gracefully/state"
)

// Check is the smallest unit of health in gracefully. Checks should be JSON
// serializable.
type Check interface {
	GetMetadata() Metadata
	Watch(ctx context.Context, channel chan Report)
}

// Metadata contains information common to every check.
type Metadata struct {
	Name    string `json:"name"`
	Runbook string `json:"runbook,omitempty"`
	Weight  uint   `json:"weight"`
}

// Result represents the outcome of a given check. This information is useful
// to help diagnose issues in the system.
type Result struct {
	State     state.State `json:"state"`
	CurrentHP   float32 `json:"currentHP,omitempty"`
	Error     error       `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}
