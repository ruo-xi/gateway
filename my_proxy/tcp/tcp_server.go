package tcp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type contextKey struct {
	value string
}

func (k contextKey) String() {
	fmt.Println(k.value)
}

var (
	ErrServerClosed     = errors.New("tcp: Server closed")
	ErrAbortHandler     = errors.New("tcp: abort TCPHandler")
	ServerContextKey    = &contextKey{"tcp-server"}
	LocalAddrContextKey = &contextKey{"local-addr"}
)

type onceCloseListener struct {
	net.Listener
	once     sync.Once
	closeErr error
}

func (l *onceCloseListener) Close() error {
	l.once.Do(l.close)
	return l.closeErr
}

func (l *onceCloseListener) close() {
	l.closeErr = l.Listener.Close()
}

type TCPHandler interface {
	ServeTCP(ctx context.Context, conn net.Conn)
}
type TcpServer struct {
	Addr    string
	Handler TCPHandler

	WriteTimeout     time.Duration
	ReadTimeout      time.Duration
	KeepAliveTimeout time.Duration

	isShutDown int32
	downChan   chan struct{}
	l          *onceCloseListener
	mu         sync.Mutex
	BaseCtx    context.Context
}

func (s *TcpServer) ListenAndServe() error {
	//首先判断服务是否关闭
	if s.shuttingDown() == true {
		return ErrServerClosed
	}
	if s.downChan == nil {
		s.downChan = make(chan struct{})
	}
	addr := s.Addr
	if addr == "" {
		return errors.New("needed address")
	}

	l, err := net.Listen("tcp", addr)

	if err != nil {
		return err
	}

	return s.Serve(l.(*net.TCPListener))
}

func (s *TcpServer) Serve(l net.Listener) error {
	s.l = &onceCloseListener{
		Listener: l,
	}
	defer s.l.Close()
	if s.BaseCtx == nil {
		s.BaseCtx = context.Background()
	}
	baseCtx := s.BaseCtx
	ctx := context.WithValue(baseCtx, ServerContextKey, s)
	for {
		rwc, err := l.Accept()
		if err != nil {
			select {
			case <-s.getDoneChan():
				return ErrServerClosed
			default:
			}
			fmt.Println("accept fail, err ", err)
			continue
		}
		c := s.newConn(rwc)
		go c.serve(ctx)
	}
	return nil
}

func (s *TcpServer) Close() error {
	atomic.StoreInt32(&s.isShutDown, 1)
	close(s.downChan)
	s.l.Close()
	return s.l.closeErr
}

func (s *TcpServer) getDoneChan() <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.downChan == nil {
		return make(chan struct{})
	}
	return s.downChan
}
func (s *TcpServer) shuttingDown() bool {
	return atomic.LoadInt32(&s.isShutDown) != 0
}

func (s *TcpServer) newConn(c net.Conn) *conn {
	newConn := &conn{
		server: s,
		rwc:    c,
	}
	if d := s.WriteTimeout; d != 0 {
		newConn.rwc.SetWriteDeadline(time.Now().Add(s.WriteTimeout))
	}
	if d := s.ReadTimeout; d != 0 {
		newConn.rwc.SetReadDeadline(time.Now().Add(s.ReadTimeout))
	}

	if d := s.KeepAliveTimeout; d != 0 {
		if tcpConn, ok := c.(*net.TCPConn); ok {
			tcpConn.SetKeepAlive(true)
			tcpConn.SetKeepAlivePeriod(d)
		}
		newConn.rwc.SetDeadline(time.Now().Add(s.KeepAliveTimeout))
	}
	return newConn
}
