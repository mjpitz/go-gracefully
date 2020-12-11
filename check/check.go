package check

import (
	"context"
	"encoding/json"
	"fmt"
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
	CurrentHP float32     `json:"currentHP,omitempty"`
	Error     error       `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// WrapError will wrap the supplied err (if present) with a JSON serializable wrapper.
func WrapError(err error) *Error {
	if err == nil {
		return nil
	}
	return &Error{err}
}

// Error is an error that is JSON serializable.
type Error struct {
	error
}

var _ error = &Error{}

// Unwrap provides access to the underlying error object.
func (e *Error) Unwrap() error {
	return e.error
}

// MarshalJSON enables JSON serialization of this error.
func (e *Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.error.Error())
}

// UnmarshalJSON enables JSON deserialization of this error.
func (e *Error) UnmarshalJSON(bytes []byte) error {
	message := ""
	err := json.Unmarshal(bytes, &message)
	if err != nil {
		return err
	}

	e.error = fmt.Errorf(message)
	return nil
}

var _ json.Marshaler = &Error{}
var _ json.Unmarshaler = &Error{}
