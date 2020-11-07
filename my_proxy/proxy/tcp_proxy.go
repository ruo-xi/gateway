package proxy

import (
	"context"
	"io"
	"log"
	"net"
	"time"
)

func NewTcpLoadBalanceReverseProxy() *TcpReverseProxy {
	return &TcpReverseProxy{
		Addr:            "127.0.0.1:6379",
	}
}

type TcpReverseProxy struct {
	Addr string

	DiaoCtx         context.Context
	DialTimeout     time.Duration
	KeepAlivePeriod time.Duration
	DialContext     func(ctx context.Context, network string, address string) (net.Conn, error)

	OnDialError func(src net.Conn, dstDialErr error)
}

func (t *TcpReverseProxy) dialContext() func(ctx context.Context, network string, address string) (net.Conn, error) {
	if t.DialContext != nil {
		return t.DialContext
	}
	return (&net.Dialer{
		Timeout:   t.DialTimeout,
		KeepAlive: t.KeepAlivePeriod,
	}).DialContext
}

func (dp *TcpReverseProxy) onDialError() func(src net.Conn, dstDialErr error) {
	if dp.OnDialError != nil {
		return dp.OnDialError
	}
	return func(src net.Conn, dstDialErr error) {
		log.Printf("tcpproxy: for incoming conn %v, error dialing %q: %v", src.RemoteAddr().String(), dp.Addr, dstDialErr)
		src.Close()
	}
}

func (t *TcpReverseProxy) ServeTCP(ctx context.Context, conn net.Conn) {
	var cancel context.CancelFunc
	if t.DialTimeout >= 0 {
		ctx, cancel = context.WithTimeout(ctx, t.dialTimeout())
	}
	dst, err := t.dialContext()(ctx, "tcp", t.Addr)
	if cancel != nil {
		cancel()
	}
	if err != nil {
		t.onDialError()(conn, err)
		log.Fatal(err)
		return
	}

	if ka := t.keepAlivePeriod(); ka > 0 {
		if c, ok := dst.(*net.TCPConn); ok {
			c.SetKeepAlive(true)
			c.SetKeepAlivePeriod(ka)
		}

	}
	defer dst.Close()
	errc := make(chan error, 1)
	go copyConn(errc, dst, conn)
	go copyConn(errc, conn, dst)

	<-errc
}

func copyConn(errc chan error, dst net.Conn, conn net.Conn) {
	_, err := io.Copy(dst, conn)
	errc <- err
}

func (t *TcpReverseProxy) dialTimeout() time.Duration {
	if t.DialTimeout == 0 {
		t.DialTimeout = time.Second * 10
	}
	return t.DialTimeout
}

func (t *TcpReverseProxy) keepAlivePeriod() time.Duration {
	if t.KeepAlivePeriod != 0 {
		return t.KeepAlivePeriod
	}
	return time.Minute
}
