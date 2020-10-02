package health

import (
	"encoding/json"
	"net/http"

	"github.com/mjpitz/go-gracefully/state"
)

// HandlerFunc returns an http.HandlerFunc for users to register with their system.
func HandlerFunc(monitor *Monitor) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		report := monitor.Report()
		body, err := json.Marshal(report)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		if report.State == state.Outage {
			writer.WriteHeader(http.StatusInternalServerError)
		} else {
			writer.WriteHeader(http.StatusOK)
		}

		_, _ = writer.Write(body)
	}
}
