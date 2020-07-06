package state_test

import (
	"testing"

	"github.com/mjpitz/go-gracefully/state"

	"github.com/stretchr/testify/require"
)

func TestRepresentation(t *testing.T) {
	require.Equal(t, "unknown", string(state.Unknown))
	require.Equal(t, "outage", string(state.Outage))
	require.Equal(t, "major", string(state.Major))
	require.Equal(t, "minor", string(state.Minor))
	require.Equal(t, "ok", string(state.OK))
}

func TestScore(t *testing.T) {
	require.Equal(t, float32(0), state.Score(state.Unknown))
	require.Equal(t, float32(0.25), state.Score(state.Outage))
	require.Equal(t, float32(0.50), state.Score(state.Major))
	require.Equal(t, float32(0.75), state.Score(state.Minor))
	require.Equal(t, float32(1), state.Score(state.OK))
}

func TestForScore(t *testing.T) {
	require.Equal(t, state.Outage, state.ForScore(0.25))
	require.Equal(t, state.Major, state.ForScore(0.50))
	require.Equal(t, state.Minor, state.ForScore(0.75))
	require.Equal(t, state.OK, state.ForScore(1))
}

func TestIdentity(t *testing.T) {
	require.Equal(t, state.Outage, state.ForScore(state.Score(state.Outage)))
	require.Equal(t, state.Major, state.ForScore(state.Score(state.Major)))
	require.Equal(t, state.Minor, state.ForScore(state.Score(state.Minor)))
	require.Equal(t, state.OK, state.ForScore(state.Score(state.OK)))
}
