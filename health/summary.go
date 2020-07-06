package health

import (
	"sync"

	"github.com/google/uuid"

	"github.com/jonboulle/clockwork"

	"github.com/mjpitz/go-gracefully/check"
)

type summary struct {
	clock       clockwork.Clock
	mu          *sync.Mutex
	checks      []check.Check
	subscribers map[string]chan *check.Report
}

func (s *summary) update(report *check.Report) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// TODO: update in memory state.

	// TODO: only broadcast when state changes

	// TODO: extra broadcast when a single report changes
	//       the overall system health

	s.broadcast(report)
}

func (s *summary) broadcast(report *check.Report) {
	for _, subscriber := range s.subscribers {
		subscriber <- report
	}
}

func (s *summary) subscribe() (chan *check.Report, UnsubFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()

	uid := uuid.New().String()
	subscriber := make(chan *check.Report, len(s.checks))

	s.subscribers[uid] = subscriber

	return subscriber, func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		delete(s.subscribers, uid)
		close(subscriber)
	}
}
