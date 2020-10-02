package health

import (
	"sync"

	"github.com/google/uuid"

	"github.com/jonboulle/clockwork"

	"github.com/mjpitz/go-gracefully/check"
	"github.com/mjpitz/go-gracefully/report"
	"github.com/mjpitz/go-gracefully/state"
)

type summary struct {
	clock clockwork.Clock
	mu    *sync.Mutex

	// health
	totalHP float32
	hp      float32

	// state
	system           *check.Result
	checks           map[string]check.Check
	lastResults      map[string]*check.Result
	lastKnownResults map[string]*check.Result
	subscribers      map[string]chan check.Report
}

func (s *summary) update(report check.Report) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// update internal data representation

	meta := report.Check.GetMetadata()

	lastResult := s.lastResults[meta.Name]
	if lastResult != nil {
		lastScore := state.Score(lastResult.State)
		s.hp -= lastScore * float32(meta.Weight)

		if lastResult.State != state.Unknown {
			s.lastKnownResults[meta.Name] = lastResult
		}
	}

	newResult := report.Result
	newScore := state.Score(newResult.State)
	s.hp += newScore * float32(meta.Weight)
	s.lastResults[meta.Name] = &newResult

	// broadcast the report if the state for the dependency changed

	if lastResult == nil || lastResult.State != newResult.State {
		s.broadcast(report)
	}

	// update system state and broadcast if it changed

	newState := state.ForScore(s.hp / s.totalHP)
	if newState != s.system.State {
		s.system.State = newState
		s.system.Timestamp = s.clock.Now()
		s.system.CurrentHP = s.hp / s.totalHP
		
		s.broadcast(check.Report{
			Result: check.Result{
				State:     s.system.State,
				Timestamp: s.system.Timestamp,
			},
		})
	}
}

func (s *summary) broadcast(report check.Report) {
	for _, subscriber := range s.subscribers {
		subscriber <- report
	}
}

func (s *summary) subscribe() (chan check.Report, UnsubFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()

	uid := uuid.New().String()

	// +1 for the system
	subscriber := make(chan check.Report, len(s.checks)+1)

	s.subscribers[uid] = subscriber

	return subscriber, func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		delete(s.subscribers, uid)
		close(subscriber)
	}
}

func (s *summary) report() report.Report {
	results := make(map[string]report.CheckResult, len(s.checks))

	for name, chk := range s.checks {
		lastResult, ok := s.lastResults[name]
		if !ok {
			lastResult = &check.Result{
				State: state.Unknown,
			}
		}

		lastKnownResult, ok := s.lastKnownResults[name]
		if !ok {
			lastKnownResult = lastResult
		}

		results[name] = report.CheckResult{
			Metadata:       chk.GetMetadata(),
			LastCheck:      *lastResult,
			LastKnownCheck: *lastKnownResult,
		}
	}

	return report.Report{
		Result:  *(s.system),
		Results: results,
	}
}
