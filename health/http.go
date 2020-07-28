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
			writer.WriteHeader(500)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		if report.State == state.Outage {
			writer.WriteHeader(500)
		} else {
			writer.WriteHeader(200)
		}

		_, _ = writer.Write(body)
	}
}
