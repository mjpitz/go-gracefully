package main

import (
    "context"
    //"http"
    "net/http"
    "log"
    "time"
    "google.golang.org/grpc"
    pb "github.com/alonsopf/sil"
    
    
    "github.com/mjpitz/go-gracefully/check"
    "github.com/mjpitz/go-gracefully/health"
    "github.com/mjpitz/go-gracefully/state"
)

const (
    address     = "localhost:50051"
    defaultName = "world"
)
var c pb.GreeterClient
func main() {
    var err error
    var conn *grpc.ClientConn
    conn, err = grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
    if err != nil {
        log.Fatalf("did not connect: %v", err)
    }
    defer conn.Close()
    //don't forget to run the server first
    c = pb.NewGreeterClient(conn)
    
    monitor := health.NewMonitor([]check.Check{
        &check.Periodic{
            Metadata: check.Metadata{
                Name: "health_check",
                //Runbook: "https://example.com/Runbook.md",
                Weight: 10,
            },
            Interval: time.Second * 5,
            Timeout: time.Second,
            RunFunc: func(ctx context.Context) (state.State, error) {
                // make GRPC call, check system health
                ctx, cancel := context.WithTimeout(context.Background(), time.Second)
                defer cancel()
                _, err := c.SayHello(ctx, &pb.HelloRequest{Name: "hello grpc"})
                if err != nil {
                    return state.Unknown, err
                }
                return state.OK, nil
            },
        },
        &check.Periodic{
            Metadata: check.Metadata{
                Name: "check-if-grpc-is-up",
                //Runbook: "https://example.com/Runbook.md",
                Weight: 15,
            },
            Interval: time.Second * 5,
            Timeout: time.Second,
            RunFunc: func(ctx context.Context) (state.State, error) {
                // make GRPC call, check system health
                ctx, cancel := context.WithTimeout(context.Background(), time.Second)
                defer cancel()
                _, err := c.SayHelloAgain(ctx, &pb.HelloRequest{Name: "hello grpc again"})
                if err != nil {
                    return state.Unknown, err
                }
                return state.OK, nil
            },
        },
    }...)

    //reports, unsubscribe := monitor.Subscribe()
    //defer unsubscribe()

    ctx := context.Background()
    if err := monitor.Start(ctx); err != nil {
        log.Fatal(err.Error())
    }
/*
    // subscribe to changes in health
    for report := range reports {
        // access check information if present
        // - Check will not be present for changes in overall system health
        _ = report.Check

        // access check evaluation result data
        // - Result will be present for all reports
        _ = report.Result
    }
  */  
    // or add an HTTP endpoint to view the results of it
    http.HandleFunc("/healthz", health.HandlerFunc(monitor))
    if err := http.ListenAndServe(":9973", nil); err != nil {
        log.Fatal(err)
    }
    //fmt.Println("listen and serve in http://sil.red:9973/healthz")
}