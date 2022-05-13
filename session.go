package main

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/go-logr/logr"
)

// chain 可以有很多 session

// udp 会话是  frontend <-> backend 之间传递

type udpSession struct {
	AliveTime atomic.Value
	TTL       time.Duration
	Ch        chan []byte
}

// udpSessionBackend 是右边的会话
type udpSessionBackend struct {
	ParentSsn *udpSession
	ToCnn     net.Conn
}

// errors.Is(err, syscall.ECONNREFUSED) {
// If peer not UDP listen, then we will recv ICMP
// golang will give error net.OpError ECONNREFUSED
// we ignore this error

func (u udpSessionBackend) Close() error {
	return u.ToCnn.Close()
}

// Read 读取 udp 上的数据，如果超时了 要判断是否这个链路上没有数据存活
func (u udpSessionBackend) Read(p []byte) (int, error) {
	for {
		// add one more second to make sure  now - alive time > ttl
		u.ToCnn.SetReadDeadline(time.Now().Add(u.ParentSsn.TTL).Add(time.Second))
		n, err := u.ToCnn.Read(p)
		u.ToCnn.SetReadDeadline(time.Time{})
		if err == nil {
			u.ParentSsn.AliveTime.Store(time.Now())
			return n, err
		}

		whenAliveVoid := u.ParentSsn.AliveTime.Load()
		if whenAliveVoid == nil {
			return 0, err
		}
		// ttl arrived
		whenAlive := whenAliveVoid.(time.Time)
		now := time.Now()
		if now.After(whenAlive) && now.Sub(whenAlive) > u.ParentSsn.TTL {
			return 0, err
		}
	}
}

func (u udpSessionBackend) Write(p []byte) (int, error) {
	n, err := u.ToCnn.Write(p)
	u.ParentSsn.AliveTime.Store(time.Now())
	return n, err
}

// udpSessionFrontend 是左边的会话
type udpSessionFrontend struct {
	PktCnn net.PacketConn
	Addr   net.Addr
	Ch     chan []byte
	Closed chan struct{}
}

func (u *udpSessionFrontend) Close() error {
	select {
	case <-u.Closed:
	default:
		close(u.Closed)
	}
	return nil
}

func (u *udpSessionFrontend) Read(p []byte) (int, error) {
	select {
	case <-u.Closed:
		return 0, fmt.Errorf("udp session frontend is closed")
	case body := <-u.Ch:
		n := copy(p, body)
		return n, nil
	}
}

func (u *udpSessionFrontend) Write(p []byte) (int, error) {
	return u.PktCnn.WriteTo(p, u.Addr)
}

func handleTcpSession(waitCtx context.Context, logger logr.Logger, c *chain, cnn net.Conn) {
	var err error
	var toCnn net.Conn

	// connect peer
	d := &net.Dialer{}
	if toCnn, err = d.DialContext(waitCtx, c.Proto, c.ToAddr); err != nil {
		logger.Error(err, "failed dial tcp addr", "addr", c.ToAddr)
		return
	}
	logger = logger.WithValues("left", fmt.Sprintf("%v to %v", cnn.RemoteAddr().String(), cnn.LocalAddr().String()),
		"right", fmt.Sprintf("%v to %v", toCnn.LocalAddr().String(), toCnn.RemoteAddr().String()))
	logger.Info("new connection pair")
	defer logger.Info("close connection pair")
	readWriteEach(waitCtx, cnn, toCnn)
}

func handleUdpSession(waitCtx context.Context, logger logr.Logger, c *chain, ssn *udpSession,
	frontend *udpSessionFrontend, body []byte) {
	var err error
	var toCnn net.Conn

	d := &net.Dialer{}
	if toCnn, err = d.DialContext(waitCtx, c.Proto, c.ToAddr); err != nil {
		logger.Error(err, "failed dial udp addr", "addr", c.ToAddr)
		return
	}
	logger = logger.WithValues("left", fmt.Sprintf("%v to %v", frontend.Addr.String(), frontend.PktCnn.LocalAddr().String()),
		"right", fmt.Sprintf("%v to %v", toCnn.LocalAddr().String(), toCnn.RemoteAddr().String()))
	logger.Info("new connection pair")
	defer logger.Info("close connection pair")

	backend := udpSessionBackend{
		ParentSsn: ssn,
		ToCnn:     toCnn,
	}

	if _, err = backend.Write(body); err != nil {
		logger.Error(err, "failed write on udp")
		return
	}
	readWriteEach(waitCtx, frontend, backend)
}
