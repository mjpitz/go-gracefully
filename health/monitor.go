package health

import (
	"context"
	"sync"

	"github.com/jonboulle/clockwork"

	"github.com/mjpitz/go-gracefully/check"
	"github.com/mjpitz/go-gracefully/state"
)

// NewMonitor constructs and returns a monitor capable of observing the
// provided set of checks. Monitors must be started and can be observed.
func NewMonitor(checks ...check.Check) *Monitor {
	clock := clockwork.NewRealClock()

	totalHP := float32(0)
	checkIndex := make(map[string]check.Check)
	for _, registered := range checks {
		meta := registered.GetMetadata()
		checkIndex[meta.Name] = registered
		totalHP += float32(meta.Weight)
	}

	return &Monitor{
		clock:   clock,
		mu:      &sync.Mutex{},
		started: false,
		summary: &summary{
			clock:       clock,
			mu:          &sync.Mutex{},
			checks:      checkIndex,
			subscribers: make(map[string]chan *check.Report),
			totalHP:     totalHP,
			hp:          0,
			system: &check.Result{
				State: state.Unknown,
			},
			lastResults:      make(map[string]*check.Result),
			lastKnownResults: make(map[string]*check.Result),
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
		// +1 for the system
		reports := make(chan *check.Report, len(m.summary.checks) + 1)

		for _, registered := range m.summary.checks {
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
