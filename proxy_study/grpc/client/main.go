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
	"sync"
	"time"
)

var addr = flag.String("addr", "localhost:50051", "the address to connect to")

const message = "this is examples/metadata"
const (
	timestampFormat = time.UnixDate
	streamingCount  = 10
	AccessToken     = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1ODk2OTExMTQsImlzcyI6ImFwcF9pZF9iIn0.qb2A_WsDP_-jfQBxJk6L57gTnAzZs-SPLMSS_UO6Gkc"
)

func main() {
	flag.Parse()
	wg := sync.WaitGroup{}
	for i := 0; i < 1; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := grpc.Dial(*addr, grpc.WithInsecure())
			if err != nil {
				log.Fatal(err)
			}

			defer conn.Close()

			c := proto.NewEchoClient(conn)

			unaryCallWithMetadata(c, message)
			time.Sleep(time.Millisecond * 400)

			serverStreamingWithMetadata(c, message)
			time.Sleep(time.Millisecond * 400)

			clientStramingWithMetadata(c, message)
			time.Sleep(time.Millisecond * 400)

			bitrectionalWithMetadata(c, message)
			time.Sleep(time.Millisecond * 400)
		}()
		wg.Wait()
	}
}

func bitrectionalWithMetadata(c proto.EchoClient, m string) {
	fmt.Println("--- bitdirectional ---")
	md := metadata.Pairs("timestamp", time.Now().Format(timestampFormat))
	md.Append("Authorization", "Bearer"+AccessToken)
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	stream, err := c.BidirectionalStreamingEcho(ctx)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for i := 0; i < streamingCount; i++ {
			if err := stream.Send(&proto.EchoRequest{Message: m}); err != nil {
				log.Println(err)
			}
		}
		stream.CloseSend()
	}()

	var rpcStatus error
	for true {
		r, err := stream.Recv()
		if err != nil {
			rpcStatus = err
			break
		}
		fmt.Printf("- %s\n", r.Message)
	}
	if rpcStatus != io.EOF {
		log.Println(rpcStatus)
	}
}

func clientStramingWithMetadata(c proto.EchoClient, m string) {
	fmt.Println("--- client streaming ---")
	md := metadata.Pairs("timestamp", time.Now().Format(timestampFormat))
	md.Append("authorization", "Bearer "+AccessToken)
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	stream, err := c.ClientStreamingEcho(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < streamingCount; i++ {
		err := stream.Send(&proto.EchoRequest{Message: m})
		if err != nil {
			log.Println(err)
		}
	}

	r, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("response:%v\n", r.Message)

}

func serverStreamingWithMetadata(c proto.EchoClient, m string) {
	fmt.Println("--- server streaming ---")
	md := metadata.Pairs("timestamp", time.Now().Format(timestampFormat))
	md.Append("authorization", "Bearer "+AccessToken)
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	stream, err := c.ServerStreamingEcho(ctx, &proto.EchoRequest{Message: m})
	if err != nil {
		log.Fatal(err)
	}
	var rpcStatus error
	for {
		r, err := stream.Recv()
		if err != nil {
			rpcStatus = err
			break
		}
		fmt.Printf("- %s\n", r.Message)
	}
	if rpcStatus != io.EOF {
		log.Fatal(rpcStatus)
	}
}

func unaryCallWithMetadata(c proto.EchoClient, m string) {
	fmt.Println("--- unary ---")
	md := metadata.Pairs("timestamp", time.Now().Format(timestampFormat))
	md.Append("authorization", "Bearer "+AccessToken)
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	r, err := c.UnaryEcho(ctx, &proto.EchoRequest{Message: m})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("response:%v\n\n", r.Message)
}
