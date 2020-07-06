package health

import (
	"context"
	"sync"

	"github.com/jonboulle/clockwork"

	"github.com/mjpitz/go-gracefully/check"
)

// NewMonitor constructs and returns a monitor capable of observing the
// provided set of checks. Monitors must be started and can be observed.
func NewMonitor(checks ...check.Check) *Monitor {
	clock := clockwork.NewRealClock()

	return &Monitor{
		clock:   clock,
		mu:      &sync.Mutex{},
		started: false,
		checks:  checks,
		summary: &summary{
			clock:       clock,
			mu:          &sync.Mutex{},
			checks:      nil,
			subscribers: nil,
		},
	}
}

// Monitor is process used to manage the registered checks. It's responsible
// for consolidating all check reports from each check and broadcasting them
// out to subscribers.
type Monitor struct {
	clock clockwork.Clock

	mu      *sync.Mutex
	started bool
	checks  []check.Check
	summary *summary
}

// SetClock updates the internal clock used by the system. This must be called
// before the system is started. Once started, the clock cannot be changed.
func (m *Monitor) SetClock(clock clockwork.Clock) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		return ErrAlreadyStarted
	}

	m.clock = clock
	m.summary.clock = clock

	return nil
}

// Start initiates all check watches.
func (m *Monitor) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		return ErrAlreadyStarted
	}

	m.started = true

	go func() {
		reports := make(chan *check.Report, len(m.checks))

		for _, registered := range m.checks {
			registered.Watch(ctx, reports)
		}

		stopCh := ctx.Done()
		for {
			select {
			case report := <-reports:
				m.summary.update(report)
			case <-stopCh:
				return
			}
		}
	}()

	return nil
}

// Subscribe returns a channel that buffers reports for subscribers to respond to.
// A report who has no check specified represents a change in overall system health.
func (m *Monitor) Subscribe() (chan *check.Report, UnsubFunc) {
	return m.summary.subscribe()
}
