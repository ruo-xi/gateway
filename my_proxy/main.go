package main

import (
	"gateway/my_proxy/proxy"
	"gateway/my_proxy/tcp"
	"gateway/my_proxy/tcp_middleware"
)

func main() {
	tcpSliceRouter := tcp_middleware.NewTcpSliceRouter()
	handler := tcp_middleware.NewTcpSliceRouterHandler(tcpSliceRouter, func(context *tcp_middleware.TcpSliceRouterContext) tcp.TCPHandler {
		return proxy.NewTcpLoadBalanceReverseProxy()
	})
	server := &tcp.TcpServer{
		Addr:    ":1234",
		Handler: handler,
	}
	server.ListenAndServe()
}
