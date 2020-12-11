package check_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"

	"github.com/mjpitz/go-gracefully/check"
	"github.com/mjpitz/go-gracefully/state"

	"github.com/stretchr/testify/require"
)

func TestPeriodic_Once_Healthy(t *testing.T) {
	healthy := &check.Periodic{
		Timeout: time.Second * 10,
		Clock:   clockwork.NewFakeClock(),
		RunFunc: func(ctx context.Context) (state.State, error) {
			return state.OK, nil
		},
	}

	result := healthy.Once(context.TODO())
	require.Equal(t, state.OK, result.State)
	require.Nil(t, result.Error)
}

func TestPeriodic_Once_Unhealthy(t *testing.T) {
	unhealthy := &check.Periodic{
		Timeout: time.Second * 10,
		Clock:   clockwork.NewFakeClock(),
		RunFunc: func(ctx context.Context) (state.State, error) {
			return state.Outage, fmt.Errorf("failure")
		},
	}

	result := unhealthy.Once(context.TODO())
	require.Equal(t, state.Outage, result.State)
	require.NotNil(t, result.Error)
	require.Equal(t, "failure", result.Error.Error())
}

func TestPeriodic_Once_Timeout(t *testing.T) {
	timeout := &check.Periodic{
		// don't set a timeout to trigger timeout error
		Clock: clockwork.NewFakeClock(),
		RunFunc: func(ctx context.Context) (state.State, error) {
			return state.OK, nil
		},
	}

	result := timeout.Once(context.TODO())
	require.Equal(t, state.Unknown, result.State)
	require.NotNil(t, result.Error)

	wrapped := result.Error.(*check.Error).Unwrap()
	require.Equal(t, check.ErrTimeout, wrapped)
}

func TestPeriodic_Watch(t *testing.T) {
	clock := clockwork.NewFakeClock()

	responses := []state.State{
		state.Outage,
		state.Major,
		state.Minor,
		state.OK,
	}

	count := 0

	periodic := &check.Periodic{
		Interval: time.Second,
		Timeout:  time.Second * 10,
		Clock:    clock,
		RunFunc: func(ctx context.Context) (state.State, error) {
			response := responses[count]
			count++
			return response, nil
		},
	}

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	reportChan := make(chan check.Report, 1)

	periodic.Watch(ctx, reportChan)

	{
		report := <-reportChan
		require.Equal(t, state.Outage, report.Result.State)
		require.Nil(t, report.Result.Error)
	}

	clock.Advance(time.Second)

	{
		report := <-reportChan
		require.Equal(t, state.Major, report.Result.State)
		require.Nil(t, report.Result.Error)
	}

	clock.Advance(time.Second)

	{
		report := <-reportChan
		require.Equal(t, state.Minor, report.Result.State)
		require.Nil(t, report.Result.Error)
	}

	clock.Advance(time.Second)

	{
		report := <-reportChan
		require.Equal(t, state.OK, report.Result.State)
		require.Nil(t, report.Result.Error)
	}
}
