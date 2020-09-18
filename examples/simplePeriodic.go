package main

import (
    "context"
    //"http"
    "net/http"
    "log"
    "time"
    "github.com/mjpitz/go-gracefully/check"
    "github.com/mjpitz/go-gracefully/health"
    "github.com/mjpitz/go-gracefully/state"
)

func main() {
    monitor := health.NewMonitor([]check.Check{
        &check.Periodic{
            Metadata: check.Metadata{
                Name: "health_check",
                //Runbook: "https://yourURL/Runbook.md",
                Weight: 10,
            },
            Interval: time.Second * 5,
            Timeout: time.Second,
            RunFunc: func(ctx context.Context) (state.State, error) {
                // make API call, check system health
                resp, err := http.Get("https://yourURL/health_check")
                if err != nil {
                    return state.Unknown, err
                }
                if resp.StatusCode == 200 {
                    return state.OK, nil
                }
                if resp.StatusCode >= 201 && resp.StatusCode <= 299 {
                    return state.Minor, err
                }
                if resp.StatusCode >= 300 && resp.StatusCode <= 399 {
                    return state.Major, err
                }
                if resp.StatusCode >= 400 && resp.StatusCode <= 499 {
                    return state.Outage, err
                }
                /*if resp.StatusCode >= 500 {
                    return state.Unknown, err
                } */       
                return state.Unknown, err
            },
        },
        &check.Periodic{
            Metadata: check.Metadata{
                Name: "check-if-grpc-is-up",
                //Runbook: "https://yourURL/Runbook.md",
                Weight: 15,
            },
            Interval: time.Second * 5,
            Timeout: time.Second,
            RunFunc: func(ctx context.Context) (state.State, error) {
                // make API call, check if grpc is up
                resp, err := http.Get("https://yourURL/checkGRPC")
                if err != nil {
                    return state.Unknown, err
                }
                if resp.StatusCode == 200 {
                    return state.OK, nil
                }
                if resp.StatusCode >= 201 && resp.StatusCode <= 299 {
                    return state.Minor, err
                }
                if resp.StatusCode >= 300 && resp.StatusCode <= 399 {
                    return state.Major, err
                }
                if resp.StatusCode >= 400 && resp.StatusCode <= 499 {
                    return state.Outage, err
                }
                /*if resp.StatusCode >= 500 {
                    return state.Unknown, err
                } */       
                return state.Unknown, err
            },
        },
    }...)

    ctx := context.Background()
    if err := monitor.Start(ctx); err != nil {
        log.Fatal(err.Error())
    }

    //  add an HTTP endpoint to view the results of it
    http.HandleFunc("/healthz", health.HandlerFunc(monitor))
    if err := http.ListenAndServe(":9973", nil); err != nil {
        log.Fatal(err)
    }
    //fmt.Println("listen and serve in :9973/healthz")
}