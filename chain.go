package main

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	cerrors "github.com/cockroachdb/errors"
	"github.com/go-logr/logr"
)

func setupTcpChain(waitCtx context.Context, logger logr.Logger, c *chain) error {
	var err error
	var lc net.ListenConfig
	var sn net.Listener
	var cnn net.Conn
	var wg = &sync.WaitGroup{}

	if sn, err = lc.Listen(waitCtx, c.Proto, c.ListenAddr); err != nil {
		return cerrors.CombineErrors(fmt.Errorf("failed listen tcp addr '%s' of '%s'", c.ListenAddr, c.String()), err)
	}
	closeCh := closeWhenContext(waitCtx, sn)
	for {
		if cnn, err = sn.Accept(); err != nil {
			err = cerrors.CombineErrors(fmt.Errorf("failed call Accept()"), err)
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
		return cerrors.CombineErrors(fmt.Errorf("failed listen udp addr '%s' of '%s'", c.ListenAddr, c.String()), err)
	}
	closeCh := closeWhenContext(waitCtx, pktCnn)
	buf = make([]byte, 128*1024)
	for {
		if size, addr, err = pktCnn.ReadFrom(buf); err != nil {
			err = cerrors.CombineErrors(fmt.Errorf("failed call ReadFrom()"), err)
			break
		}

		if tryWriteUdpSession(ssnMap, addr, buf[:size]) {
			continue
		}

		ssn := &udpSession{
			AliveTime: atomic.Value{},
			TTL:       5 * time.Minute,
			Ch:        make(chan []byte, 1000),
		}
		ssnMap.Store(addr.String(), ssn)

		wg.Add(1)
		go func(arg0 net.Addr, arg1 []byte, arg2 *udpSession) {
			atomic.AddInt64(&c.SsnCount, 1)
			defer atomic.AddInt64(&c.SsnCount, -1)
			defer wg.Done()
			frontend := &udpSessionFrontend{
				PktCnn: pktCnn,
				Addr:   arg0,
				Ch:     arg2.Ch,
				Closed: make(chan struct{}),
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
	ssn.Ch <- cloneByteSlice(body)
	return true
}

func cloneByteSlice(b []byte) []byte {
	dst := make([]byte, len(b))
	copy(dst, b)
	return dst
}

func createChainRoutine(waitCtx context.Context, logger logr.Logger, wg *sync.WaitGroup, c *chain) {
	var err error

	waitCtx, c.Cancel = context.WithCancel(waitCtx)

	f := func() {
		defer wg.Done()
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

// forward TCP/UDP from ListenAddr to ToAddr
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
