# go-gracefully

`go-gracefully` is a library for monitoring and reporting on an applications' health.
Unlike many solutions out there, `go-gracefully` uses an asynchronous, stream based check scheme.
This enables real-time checks as well as push based changes.

## How does it work?

All checks in `go-gracefully` have a common block of metadata.

* `Name` is a required descriptor of the check.
  * Avoid whitespace where possible.
* An optional `Runbook` can be provided to aid in the resolution of issues.
  * When provided, this should be a valid URL.
* Finally, a `Weight` is used to determine relative importance of check.
  * A check with `Weight: 100` has a greater impact on the system health than one with `Weight: 10`. 

```go
    // ...
    Metadata: &check.Metadata{
        Name: "periodic-check",
        Runbook: "http://path/to/runbook.md",
        Weight: 10,
    },
    // ...
```

When check's evaluated, it can return one of four possible states.

* `OK` - The check is operating as expected.
* `Minor` - The check is failing and will require attention soon.
* `Major` - The check is failing and requires attention soon.
* `Outage` - The check is failing and requires attention.

Together, the `state` and the `weight` are used to approximate an applications' health.  

## Inspirations

There are a lot of prior work out there.
It's hard to list them all.
The list below was is just a few I drew inspiration from. 

* http://github.com/indeedeng/status
* https://github.com/InVisionApp/go-health
* https://godoc.org/github.com/heptiolabs/healthcheck
* https://github.com/AppsFlyer/go-sundheit

## Installation

```bash
go get -u github.com/mjpitz/go-gravefully
```

## Usage

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/mjpitz/go-gracefully/check"
    "github.com/mjpitz/go-gracefully/health"
    "github.com/mjpitz/go-gracefully/state"
)

func main() {
    monitor := health.NewMonitor([]check.Check{
        &check.Periodic{
            Metadata: &check.Metadata{
                Name: "periodic-check",
                Runbook: "http://path/to/runbook.md",
                Weight: 10,
            },
            Interval: time.Second * 5,
            Timeout: time.Second,
            RunFunc: func(ctx context.Context) (state.State, error) {
                // make API call, check system health
                return state.OK, nil
            },
        },
        &check.Stream{
        	Metadata: &check.Metadata{
                Name: "stream-check",
                Runbook: "http://path/to/runbook.md",
                Weight: 10,
            },
            WatchFunc: func(ctx context.Context, channel chan *check.Result) {
                stopCh := ctx.Done()

                var upstreamCh chan interface{}
                // make call that fills chan

                for {
                    select {
                        case <-stopCh:
                            return
                        case _ = <-upstreamCh:
                            channel <- &check.Result{
                                State: state.OK,
                                Error: nil,
                            }
                    }
                }
            },
        },
    }...)

    reports, unsubcribe := monitor.Subscribe()
    defer unsubcribe()

    ctx := context.Background()
    if err := monitor.Start(ctx); err != nil {
        log.Fatal(err.Error())
    }

    for report := range reports {
        _ = report.Check.GetMetadata() // metadata
        _ = report.Result // check evaluation
    }
}
```
