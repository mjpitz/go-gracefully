package health

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/jonboulle/clockwork"

	"github.com/mjpitz/go-gracefully/check"
)

// NewMonitor constructs and returns a monitor capable of observing the
// provided set of checks. Monitors must be started and can be observed.
func NewMonitor(checks ...check.Check) *Monitor {
	return &Monitor{
		Clock:   clockwork.NewRealClock(),
		mu:      &sync.Mutex{},
		started: false,
		checks:  checks,
		reports: make(chan *check.Report, len(checks)),
	}
}

// Monitor is process used to manage the registered checks. It's responsible
// for consolidating all check reports from each check and broadcasting them
// out to subscribers.
type Monitor struct {
	Clock       clockwork.Clock

	mu          *sync.Mutex
	started     bool
	checks      []check.Check
	reports     chan *check.Report
	subscribers map[string]chan *check.Report
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
		for _, registered := range m.checks {
			registered.Watch(ctx, m.reports)
		}

		stopCh := ctx.Done()
		for {
			select {
			case report := <-m.reports:
				m.broadcast(report)
			case <-stopCh:
				return
			}
		}
	}()

	return nil
}

func (m *Monitor) broadcast(report *check.Report) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, subscriber := range m.subscribers {
		subscriber <- report
	}
}

// UnsubFunc defines a function call that stops the subscription process.
type UnsubFunc = func()

// Subscribe allows external actors to observe changes to the health state.
func (m *Monitor) Subscribe() (chan *check.Report, UnsubFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()

	uid := uuid.New().String()
	subscriber := make(chan *check.Report, len(m.checks))

	m.subscribers[uid] = subscriber

	return subscriber, func() {
		delete(m.subscribers, uid)
		close(subscriber)
	}
}
