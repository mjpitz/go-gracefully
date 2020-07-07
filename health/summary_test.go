package health

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"

	"github.com/mjpitz/go-gracefully/check"
	"github.com/mjpitz/go-gracefully/state"

	"github.com/stretchr/testify/require"
)

type staticCheck struct {
}

func (t *staticCheck) GetMetadata() check.Metadata {
	return check.Metadata{
		Name:   "static",
		Weight: 100,
	}
}

func (t *staticCheck) Watch(ctx context.Context, channel chan check.Report) {}

var _ check.Check = &staticCheck{}

type stateTest struct {
	in  state.State
	out []state.State
}

func TestSummary(t *testing.T) {
	tests := []stateTest{
		// each state change should receive two events: one for the check, one for the system
		// re-emissions of the same state should not propagate
		{ in: state.Outage, out: []state.State{ state.Outage, state.Outage } },
		{ in: state.Outage, out: []state.State{  } },
		{ in: state.Major, out: []state.State{ state.Major, state.Major } },
		{ in: state.Major, out: []state.State{  } },
		{ in: state.Minor, out: []state.State{ state.Minor, state.Minor } },
		{ in: state.Minor, out: []state.State{  } },
		{ in: state.OK, out: []state.State{ state.OK, state.OK } },
		{ in: state.OK, out: []state.State{  } },
	}

	clock := clockwork.NewFakeClock()

	chk := &staticCheck{}

	s := &summary{
		clock:   clock,
		mu:      &sync.Mutex{},
		totalHP: 100,
		hp:      0,
		system: &check.Result{
			State: state.Unknown,
		},
		checks: map[string]check.Check{
			chk.GetMetadata().Name: chk,
		},
		lastResults:      make(map[string]*check.Result),
		lastKnownResults: make(map[string]*check.Result),
		subscribers:      make(map[string]chan check.Report),
	}

	reports, unsub := s.subscribe()
	defer unsub()

	for _, test := range tests {
		s.update(check.Report{
			Check: chk,
			Result: check.Result{
				State: test.in,
			},
		})

		for _, expected := range test.out {
			select {
			case report := <-reports:
				require.Equal(t, expected, report.Result.State)
				continue
			case <-time.After(time.Second):
				require.Fail(t, "failed to read response back from channel")
				return
			}
		}
	}
}
