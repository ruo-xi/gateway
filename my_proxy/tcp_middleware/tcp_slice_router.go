package tcp_middleware

import (
	"context"
	"gateway/my_proxy/tcp"
	"math"
	"net"
)

const abortIndex int8 = math.MaxInt8 / 2 //最多 63 个中间件

type TcpHandlerFunc func(*TcpSliceRouterContext)

type TcpSliceRouter struct {
	handlers []TcpHandlerFunc
}

func (r *TcpSliceRouter) Use(middleware ...TcpHandlerFunc) *TcpSliceRouter {
	r.handlers = append(r.handlers, middleware...)
	return r
}

type TcpSliceRouterContext struct {
	*TcpSliceRouter
	Ctx   context.Context
	conn  net.Conn
	index int8
}

func (c *TcpSliceRouterContext) Get(key interface{}) interface{} {
	return c.Ctx.Value(key)
}

func (c *TcpSliceRouterContext) Set(key, val interface{}) {
	c.Ctx = context.WithValue(c.Ctx, key, val)
}

func (c *TcpSliceRouterContext) Reset() {
	c.index = -1
}

func (c *TcpSliceRouterContext) Next() {
	c.index++
	if c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}
}

func NewTcpSliceRouter() *TcpSliceRouter {
	return &TcpSliceRouter{}
}

func newTcpSliceRouterContext(conn net.Conn, r *TcpSliceRouter, ctx context.Context) *TcpSliceRouterContext {
	c := &TcpSliceRouterContext{conn: conn, TcpSliceRouter: r, Ctx: ctx}
	c.Reset()
	return c
}

type TcpSliceRouterHandler struct {
	Router   *TcpSliceRouter
	CoreFunc func(*TcpSliceRouterContext) tcp.TCPHandler
}

func NewTcpSliceRouterHandler(router *TcpSliceRouter, coreFunc func(*TcpSliceRouterContext) tcp.TCPHandler) *TcpSliceRouterHandler {
	return &TcpSliceRouterHandler{
		Router:   router,
		CoreFunc: coreFunc,
	}

}

func (t TcpSliceRouterHandler) ServeTCP(ctx context.Context, conn net.Conn) {
	c := newTcpSliceRouterContext(conn, t.Router, ctx)
	c.handlers = append(c.handlers, func(tcpSliceRouterContext *TcpSliceRouterContext) {
		t.CoreFunc(tcpSliceRouterContext).ServeTCP(ctx, conn)
	})
	c.Reset()
	c.Next()
}
