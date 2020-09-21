package main

import (
    "context"
    "log"
    "net/http"
    "time"

    "github.com/mjpitz/go-gracefully/check"
    "github.com/mjpitz/go-gracefully/health"
    "github.com/mjpitz/go-gracefully/state"

    "google.golang.org/grpc"
    pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
    address = "localhost:50051"
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
            Timeout:  time.Second,
            RunFunc: func(ctx context.Context) (state.State, error) {
                _, err := c.SayHello(ctx, &pb.HelloRequest{Name: "hello grpc"})
                if err != nil {
                    return state.Unknown, err
                }
                return state.OK, nil
            },
        },
    }...)

    ctx := context.Background()
    if err := monitor.Start(ctx); err != nil {
        log.Fatal(err.Error())
    }

    // add an HTTP endpoint to view the results of it
    http.HandleFunc("/healthz", health.HandlerFunc(monitor))
    if err := http.ListenAndServe(":9973", nil); err != nil {
        log.Fatal(err)
    }

}