package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-logr/logr"
)

func setupTcpChain(waitCtx context.Context, logger logr.Logger, c *chain) error {
	var err error
	var lc net.ListenConfig
	var sn net.Listener
	var cnn net.Conn
	var wg = &sync.WaitGroup{}

	if sn, err = lc.Listen(waitCtx, c.Proto, c.ListenAddr); err != nil {
		return fmt.Errorf("failed listen tcp addr '%s' of '%s' with error %w", c.ListenAddr, c.String(), err)
	}
	var closeCh = closeWhenContext(waitCtx, sn)
	for {
		if cnn, err = sn.Accept(); err != nil {
			if errors.Is(err, net.ErrClosed) {
				err = nil
			} else {
				err = fmt.Errorf("failed call Accept() with error %w", err)
			}
			break
		}
		wg.Add(1)
		go func(arg0 net.Conn) {
			atomic.AddInt64(&c.SsnCount, 1)
			defer atomic.AddInt64(&c.SsnCount, -1)
			defer wg.Done()
			handleTcpSession(waitCtx, logger, c, arg0)
		}(cnn)
	}
	close(closeCh)
	wg.Wait()
	return err
}

func setupUdpChain(waitCtx context.Context, logger logr.Logger, c *chain) error {
	var lc net.ListenConfig
	var err error
	var pktCnn net.PacketConn
	var size int
	var addr net.Addr
	var buf []byte
	var wg = &sync.WaitGroup{}
	var ssnMap = &sync.Map{}

	if pktCnn, err = lc.ListenPacket(waitCtx, c.Proto, c.ListenAddr); err != nil {
		return fmt.Errorf("failed listen udp addr '%s' of '%s' with error %w", c.ListenAddr, c.String(), err)
	}
	var closeCh = closeWhenContext(waitCtx, pktCnn)
	buf = make([]byte, 128*1024)
	for {
		if size, addr, err = pktCnn.ReadFrom(buf); err != nil {
			if errors.Is(err, net.ErrClosed) {
				err = nil
			} else {
				err = fmt.Errorf("failed call ReadFrom() with error %w", err)
			}
			break
		}

		if tryWriteUdpSession(ssnMap, addr, buf[:size]) {
			continue
		}

		var ssn = &udpSession{
			AliveTime: atomic.Value{},
			TTL:       5 * time.Minute,
			ch:        make(chan []byte, 1000),
		}
		ssnMap.Store(addr.String(), ssn)

		wg.Add(1)
		go func(arg0 net.Addr, arg1 []byte, arg2 *udpSession) {
			atomic.AddInt64(&c.SsnCount, 1)
			defer atomic.AddInt64(&c.SsnCount, -1)
			defer wg.Done()
			frontend := &udpSessionFrontend{
				pktCnn: pktCnn,
				addr:   arg0,
				ch:     arg2.ch,
				closed: make(chan struct{}),
			}
			handleUdpSession(waitCtx, logger, c, arg2, frontend, arg1)
			ssnMap.Delete(arg0.String())
		}(addr, cloneByteSlice(buf[:size]), ssn)
	}

	close(closeCh)
	wg.Wait()
	return err
}

func tryWriteUdpSession(ssnMap *sync.Map, addr net.Addr, body []byte) bool {
	ssnVoid, ok := ssnMap.Load(addr.String())
	if !ok {
		return false
	}
	ssn := ssnVoid.(*udpSession)
	ssn.Write(body)
	return true
}

func createChainRoutine(waitCtx context.Context, logger logr.Logger, wg *sync.WaitGroup, c *chain) {
	var err error

	waitCtx, c.Cancel = context.WithCancel(waitCtx)

	var f = func() {
		defer wg.Done()
		logger = logger.WithValues("chain", c.String())
		if c.Proto == "tcp" {
			if err = setupTcpChain(waitCtx, logger, c); err != nil {
				logger.Error(err, "failed create tcp chain", "chain", c.String())
			}
		}
		if c.Proto == "udp" {
			if err = setupUdpChain(waitCtx, logger, c); err != nil {
				logger.Error(err, "failed create udp chain", "chain", c.String())
			}
		}
	}
	wg.Add(1)
	go f()
}

// chain forward TCP/UDP from ListenAddr to ToAddr
type chain struct {
	ListenAddr string
	Proto      string
	ToAddr     string
	SsnCount   int64
	Cancel     context.CancelFunc
}

func (c *chain) String() string {
	return fmt.Sprintf("<%v://%v-%v>", c.Proto, c.ListenAddr, c.ToAddr)
}
