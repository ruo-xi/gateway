package main

import (
	"context"
	"fmt"
	"github.com/e421083458/grpc-proxy/proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"strings"
)

const port = ":50051"

func main() {
	l, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}
	director := func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
		if strings.HasPrefix(fullMethodName, "/com.example.internal.") {
			return ctx, nil, status.Errorf(codes.Unimplemented, "Unkown method")
		}
		c, err := grpc.DialContext(ctx, "localhost:50055", grpc.WithCodec(proxy.Codec()), grpc.WithInsecure())
		return ctx, c, err
	}
	s := grpc.NewServer(
		grpc.CustomCodec(proxy.Codec()),
		grpc.UnknownServiceHandler(proxy.TransparentHandler(director)))
	fmt.Printf("server start at %v\n", l.Addr())
	if err := s.Serve(l); err != nil {
		log.Fatal(err)
	}
}
