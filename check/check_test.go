package check_test

import (
	"encoding/json"
	"testing"

	"github.com/mjpitz/go-gracefully/check"
	"github.com/mjpitz/go-gracefully/state"

	"github.com/stretchr/testify/require"
)

const simpleResult = `{"state":"ok","timestamp":"0001-01-01T00:00:00Z"}`
const errorResult = `{"state":"outage","error":"timed out waiting for check","timestamp":"0001-01-01T00:00:00Z"}`

func TestResult(t *testing.T) {
	// test simple
	{
		simple := check.Result{
			State: state.OK,
		}

		data, err := json.Marshal(simple)
		require.NoError(t, err)

		require.Equal(t, simpleResult, string(data))
	}

	// test error serialization
	{
		withError := check.Result{
			State: state.Outage,
			Error: check.WrapError(check.ErrTimeout),
		}

		data, err := json.Marshal(withError)
		require.NoError(t, err)

		require.Equal(t, errorResult, string(data))
	}
}
