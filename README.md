![GitHub](https://img.shields.io/github/license/depscloud/gateway.svg)
![branch](https://github.com/depscloud/gateway/workflows/branch/badge.svg?branch=main)
![Google Analytics](https://www.google-analytics.com/collect?v=1&tid=UA-172921913-1&cid=555&t=event&ec=repo&ea=open&dp=go-gracefully&dt=go-gracefully)

# go-gracefully

`go-gracefully` is a library for monitoring and reporting on an applications' health.
Unlike many solutions out there, `go-gracefully` uses an asynchronous, stream based check scheme.
This enables real-time checks as well as push based changes.

**Status:**

This library is currently available as a _preview_.
I started development to support my existing work on [deps.cloud](http://github.com/depscloud).

Most health check libraries you find are pull based.
In this model, you define your check as a function that's called on some set interval.
While this is a great start to a solution, more advance techniques need to be able to push.

The flow of information in the system is as follows:

[![](https://mermaid.ink/img/eyJjb2RlIjoiZ3JhcGggTFJcbiAgQVtQZXJpb2RpY0NoZWNrXSAtLT58UmVwb3J0fCBDaChDaGFubmVsKVxuICBCW1N0cmVhbUNoZWNrXSAtLT58UmVwb3J0fCBDaFxuICBDW1N0cmVhbUNoZWNrXSAtLT58UmVwb3J0fCBDaFxuXHRDaCAtLT4gTVtNb25pdG9yXVxuICBNIC0tPnxSZXBvcnR8IFNbU3Vic2NyaWJlcnNdXG4gIE0gLS0tfG1haW50YWluc3wgc3VtbWFyeSIsIm1lcm1haWQiOnsidGhlbWUiOiJkZWZhdWx0In0sInVwZGF0ZUVkaXRvciI6ZmFsc2V9)](https://mermaid-js.github.io/mermaid-live-editor/#/edit/eyJjb2RlIjoiZ3JhcGggTFJcbiAgQVtQZXJpb2RpY0NoZWNrXSAtLT58UmVwb3J0fCBDaChDaGFubmVsKVxuICBCW1N0cmVhbUNoZWNrXSAtLT58UmVwb3J0fCBDaFxuICBDW1N0cmVhbUNoZWNrXSAtLT58UmVwb3J0fCBDaFxuXHRDaCAtLT4gTVtNb25pdG9yXVxuICBNIC0tPnxSZXBvcnR8IFNbU3Vic2NyaWJlcnNdXG4gIE0gLS0tfG1haW50YWluc3wgc3VtbWFyeSIsIm1lcm1haWQiOnsidGhlbWUiOiJkZWZhdWx0In0sInVwZGF0ZUVkaXRvciI6ZmFsc2V9)

1. Each `Check` produces a `Report`
2. `Reports` are collected through the use of a channel by the `Monitor`.
3. The monitor updates its `summary` of the system.
4. When either a check, or the system state changes, a `Report` is published to subscribers.

A snapshot of the full report can be obtained from the `Monitor`.

## How does it work?

All checks in `go-gracefully` have a common block of metadata.

* `Name` is a required descriptor of the check.
  * Avoid whitespace where possible. Preferred character set: `[a-zA-Z0-9-_.]`
* An optional `Runbook` can be provided to aid in the resolution of issues.
  * When provided, this should be a valid URL.
* Finally, a `Weight` is used to determine relative importance of check to the overall system.
  * A check with `Weight: 100` has a greater impact on the system health than one with `Weight: 10`. 

```go
    // ...
    Metadata: check.Metadata{
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

On its own, `state` represents a fractional value of health (i.e. `[0-1]`).
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
            Metadata: check.Metadata{
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
            Metadata: check.Metadata{
                Name: "stream-check",
                Runbook: "http://path/to/runbook.md",
                Weight: 10,
            },
            WatchFunc: func(ctx context.Context, channel chan check.Result) {
                stopCh := ctx.Done()

                var upstreamCh chan interface{}
                // make call that fills chan

                for {
                    select {
                        case <-stopCh:
                            return
                        case _ = <-upstreamCh:
                            channel <- check.Result{
                                State: state.OK,
                                Error: nil,
                            }
                    }
                }
            },
        },
    }...)

    reports, unsubscribe := monitor.Subscribe()
    defer unsubscribe()

    ctx := context.Background()
    if err := monitor.Start(ctx); err != nil {
        log.Fatal(err.Error())
    }

    for report := range reports {
        // access check information if present
        // - Check will not be present for changes in overall system health
        _ = report.Check

        // access check evaluation result data
        // - Result will be present for all reports
        _ = report.Result
    }
}
```
