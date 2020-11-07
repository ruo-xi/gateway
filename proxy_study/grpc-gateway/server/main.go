package main

import (
	"context"
	"flag"
	"fmt"
	"gateway/proxy_study/grpc-gateway/proto"
	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"net/http"
)

var (
	serverAddr         = ":8081"
	grpcServerEndPoint = flag.String("grpc_server-emdpoint", "localhost:50055", "gRPC server endpoint")
)

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := proto.RegisterEchoHandlerFromEndpoint(ctx, mux, *grpcServerEndPoint, opts)
	if err != nil {
		return err
	}
	return http.ListenAndServe(serverAddr, mux)
}
func main() {
	flag.Parse()
	defer glog.Flush()
	fmt.Println("server listening at", serverAddr)
	if err := run(); err != nil {
		glog.Fatal(err)
	}
}
