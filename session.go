package main

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"golang.org/x/exp/slog"
)

// chain include many session

// udp session include  frontend <-> backend

type udpSession struct {
	AliveTime atomic.Value
	TTL       time.Duration
	ch        chan []byte
}

// Write write bytes to frontend session
func (u *udpSession) Write(p []byte) (int, error) {
	u.ch <- cloneByteSlice(p)
	return len(p), nil
}

// udpSessionBackend is left session in chain
type udpSessionBackend struct {
	parentSsn *udpSession
	frontend  *udpSessionFrontend
	toCnn     net.Conn
}

func (u udpSessionBackend) Close() error {
	return u.toCnn.Close()
}

// Read read data from udp session, close the frontend and backend session when ttl arrived
func (u udpSessionBackend) Read(p []byte) (int, error) {
	for {
		// add one more second to make sure  now - alive time > ttl
		u.toCnn.SetReadDeadline(time.Now().Add(u.parentSsn.TTL).Add(time.Second))
		var n, err = u.toCnn.Read(p)
		u.toCnn.SetReadDeadline(time.Time{})

		if err == nil {
			u.parentSsn.AliveTime.Store(time.Now())
			return n, nil
		}
		if !u.continueReadWhenErr(err) {
			u.frontend.Close()
			return 0, err
		}
	}
}

func (u udpSessionBackend) continueReadWhenErr(err error) bool {
	var now = time.Now()

	var whenAliveVoid = u.parentSsn.AliveTime.Load()
	if whenAliveVoid == nil {
		return false
	}

	// return true only if ttl not arrived
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		// ttl arrived
		var whenAlive = whenAliveVoid.(time.Time)
		if now.After(whenAlive) && now.Sub(whenAlive) < u.parentSsn.TTL {
			return true
		}
	}
	return false

	// If peer not UDP listen, then we will recv ICMP
	// golang will give error net.OpError ECONNREFUSED
	// we ignore this error
	// if errors.Is(err, syscall.ECONNREFUSED) {
}

func (u udpSessionBackend) Write(p []byte) (int, error) {
	var n, err = u.toCnn.Write(p)
	u.parentSsn.AliveTime.Store(time.Now())
	return n, err
}

// udpSessionFrontend is the left session
type udpSessionFrontend struct {
	pktCnn net.PacketConn
	addr   net.Addr
	ch     chan []byte
	closed chan struct{}
}

func (u *udpSessionFrontend) Close() error {
	select {
	case <-u.closed:
	default:
		close(u.closed)
	}
	return nil
}

func (u *udpSessionFrontend) Read(p []byte) (int, error) {
	select {
	case <-u.closed:
		return 0, fmt.Errorf("udp session frontend is closed")
	case body := <-u.ch:
		n := copy(p, body)
		return n, nil
	}
}

func (u *udpSessionFrontend) Write(p []byte) (int, error) {
	return u.pktCnn.WriteTo(p, u.addr)
}

func handleTcpSession(waitCtx context.Context, logger *slog.Logger, c *chain, cnn net.Conn) {
	var err error
	var toCnn net.Conn

	// connect peer
	var d = &net.Dialer{}
	if toCnn, err = d.DialContext(waitCtx, c.Proto, c.ToAddr); err != nil {
		logger.Error("failed dial tcp addr", "addr", c.ToAddr, "error", err)
		return
	}
	logger = logger.With("left", fmt.Sprintf("%v to %v", cnn.RemoteAddr().String(), cnn.LocalAddr().String()),
		"right", fmt.Sprintf("%v to %v", toCnn.LocalAddr().String(), toCnn.RemoteAddr().String()))
	logger.Info("new tcp session pair")
	defer logger.Info("close tcp session pair")
	readWriteEach(waitCtx, cnn, toCnn)
}

func handleUdpSession(waitCtx context.Context, logger *slog.Logger, c *chain, ssn *udpSession,
	frontend *udpSessionFrontend, body []byte) {
	var err error
	var toCnn net.Conn

	var d = &net.Dialer{}
	if toCnn, err = d.DialContext(waitCtx, c.Proto, c.ToAddr); err != nil {
		logger.Error("failed dial udp addr", "addr", c.ToAddr, "error", err)
		return
	}
	logger = logger.With("left", fmt.Sprintf("%v to %v", frontend.addr.String(), frontend.pktCnn.LocalAddr().String()),
		"right", fmt.Sprintf("%v to %v", toCnn.LocalAddr().String(), toCnn.RemoteAddr().String()))
	logger.Info("new udp session pair")
	defer logger.Info("close udp session pair")

	var backend = udpSessionBackend{
		parentSsn: ssn,
		toCnn:     toCnn,
		frontend:  frontend,
	}

	if _, err = backend.Write(body); err != nil {
		logger.Error("failed write on udp", "error", err)
		return
	}
	readWriteEach(waitCtx, frontend, backend)
}

func cloneByteSlice(b []byte) []byte {
	dst := make([]byte, len(b))
	copy(dst, b)
	return dst
}
