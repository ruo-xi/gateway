package main

import (
	"context"
	"flag"
	"fmt"
	"gateway/proxy_study/grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"io"
	"log"
	"net"
)

var port = flag.Int("port", 50055, "the port to serve on")

const streamingCount = 10

type server struct {
}

func (s server) UnaryEcho(ctx context.Context, request *proto.EchoRequest) (*proto.EchoResponse, error) {
	fmt.Println("--- UnaryEcho ---")
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("miss metadata from context")
	}
	fmt.Println("md", md)
	fmt.Printf("request received : %v, sending echo\n", request)
	return &proto.EchoResponse{Message: request.Message}, nil
}

func (s server) ServerStreamingEcho(request *proto.EchoRequest, echoServer proto.Echo_ServerStreamingEchoServer) error {
	fmt.Println("--- ServerStreaningEcho ---")
	fmt.Printf("request received: %v \n", request)
	//md, _ := metadata.FromIncomingContext(echoServer.Context())
	//fmt.Println(md)
	for i := 0; i < streamingCount; i++ {
		fmt.Printf("echo message %v\n", request.Message)
		err := echoServer.Send(&proto.EchoResponse{Message: request.Message})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s server) ClientStreamingEcho(echoServer proto.Echo_ClientStreamingEchoServer) error {
	fmt.Println("--- ClientStreamingRcho ---")
	var message string
	for {
		echoRequest, err := echoServer.Recv()
		if err == io.EOF {
			fmt.Printf("echo last received message\n")
			return echoServer.SendAndClose(&proto.EchoResponse{Message: message})

		}
		message = echoRequest.Message
		fmt.Printf("request received: %v, building echo\n", echoRequest)
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Printf("request receive: %v \n")
	return nil
}

func (s server) BidirectionalStreamingEcho(echoServer proto.Echo_BidirectionalStreamingEchoServer) error {
	fmt.Printf("--- BidirectionalStreamingEcho ---\n")
	for true {
		echoRequest, err := echoServer.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		fmt.Printf("request received %v, sending echo \n", echoRequest)
		if err := echoServer.Send(&proto.EchoResponse{Message: echoRequest.Message}); err != nil {
			return err
		}

	}
	return nil

}

func main() {
	flag.Parse()
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("server listening at %v\n", l.Addr())
	s := grpc.NewServer()
	proto.RegisterEchoServer(s, &server{})
	s.Serve(l)
}
