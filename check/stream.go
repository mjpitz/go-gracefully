package check

import (
	"context"

	"github.com/jonboulle/clockwork"
)

// WatchFunc defines an easy to use function that starts a watch.
type WatchFunc = func(ctx context.Context, channel chan *Result)

// Stream is a Check implementation
type Stream struct {
	*Metadata
	Clock     clockwork.Clock `json:"-"`
	WatchFunc WatchFunc       `json:"-"`
}

// GetMetadata returns meta information about the check.
func (s *Stream) GetMetadata() *Metadata {
	return s.Metadata
}

// Watch will observe a stream of changes in a checks state.
func (s *Stream) Watch(ctx context.Context, channel chan *Report) {
	if s.Clock == nil {
		s.Clock = clockwork.NewRealClock()
	}

	results := make(chan *Result, 1)
	s.WatchFunc(ctx, results)

	stopCh := ctx.Done()

	go func() {
		for {
			select {
			case result := <-results:
				result.Timestamp = s.Clock.Now()
				channel <- &Report{
					Check:  s,
					Result: result,
				}
			case <-stopCh:
				return
			}
		}
	}()
}

var _ Check = &Stream{}
