package check

import (
	"context"
	"time"

	"github.com/jonboulle/clockwork"

	"github.com/mjpitz/go-gracefully/state"
)

// RunFunc is a simple function that defines a unary check for state.
type RunFunc = func(ctx context.Context) (state.State, error)

// Periodic is a Check implementation that runs a provided function on a set
// interval and configured timeout. This is a common type of Check.
type Periodic struct {
	Metadata
	Interval time.Duration   `json:"interval,string"`
	Timeout  time.Duration   `json:"timeout,string"`
	Clock    clockwork.Clock `json:"-"`
	RunFunc  RunFunc         `json:"-"`
}

// GetMetadata returns meta information about the check.
func (p *Periodic) GetMetadata() Metadata {
	return p.Metadata
}

// Once performs a one time evaluation of the check.
func (p *Periodic) Once(parent context.Context) Result {
	p.init()

	ctx, cancel := context.WithTimeout(parent, p.Timeout)
	defer cancel()

	result := make(chan Result, 1)

	go func() {
		computedState, err := p.RunFunc(ctx)
		result <- Result{
			State:     computedState,
			Error:     err,
			Timestamp: p.Clock.Now(),
		}
	}()

	select {
	case r := <-result:
		return r
	case <-p.Clock.After(p.Timeout):
		return Result{
			State:     state.Unknown,
			Error:     ErrTimeout,
			Timestamp: p.Clock.Now(),
		}
	}
}

// Watch sets up a go routine to run the check on the configured interval.
func (p *Periodic) Watch(ctx context.Context, channel chan Report) {
	p.init()

	stopCh := ctx.Done()

	go func() {
		for {
			result := p.Once(ctx)
			channel <- Report{
				Check:  p,
				Result: result,
			}

			select {
			case <-p.Clock.After(p.Interval):
				continue
			case <-stopCh:
				return
			}
		}
	}()
}

func (p *Periodic) init() {
	if p.Clock == nil {
		p.Clock = clockwork.NewRealClock()
	}
}

var _ Check = &Periodic{}
