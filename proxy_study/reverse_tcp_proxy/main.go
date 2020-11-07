package main

import (
	"context"
	"net"
)

var addr = ":2002"

type tcpHandler struct {
}

func (receiver *tcpHandler) ServeTCP(ctx context.Context, src net.Conn) {
	src.Write([]byte("tcpHandler\n"))
}


//type TcpServer struct {
//	l net.Listener
//}
//
//func (s *TcpServer) ListenAndServe() {
//	conn, err := net.Listen("tcp", addr)
//	if err != nil {
//		log.Fatal(err)
//	}
//}

func main() {

}
