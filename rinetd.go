package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// worked like `rinetd`.

// forward TCP/UDP from ListenAddr to ToAddr
type chain struct {
	ListenAddr string
	Proto      string
	ToAddr     string
}

func (c *chain) String() string {
	return fmt.Sprintf("%v->%v %v", c.ListenAddr, c.ToAddr, c.Proto)
}

// Keep udp session
type udpSession struct {
	FromCnn    net.PacketConn
	FromAddr   net.Addr
	OwnerChain *chain
	WriteTime  int64
	ToCnn      net.Conn
}

func (u *udpSession) Close() error {
	return u.ToCnn.Close()
}

type mgt struct {
	// 业务集合
	Chains    []*chain
	UdpSsns   sync.Map // hash udpSession
	TcpCnnCnt int64
	// stat
	StatInterval time.Duration
	UdpTTLSec    int64
	// sync
	WaitCtx context.Context
	Wg      *sync.WaitGroup
}

func (m *mgt) UdpCnnCnt() uint64 {
	r := uint64(0)
	m.UdpSsns.Range(func(key, value interface{}) bool {
		r++
		return true
	})
	return r
}

func setupSignal(mgt0 *mgt, cancel context.CancelFunc) {
	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, os.Interrupt)
	signal.Notify(sigCh, syscall.SIGTERM)
	mgt0.Wg.Add(1)
	go func() {
		select {
		case <-sigCh:
			cancel()
		case <-mgt0.WaitCtx.Done():
		}
		mgt0.Wg.Done()
	}()

}

// 可以通过 WaitCtx.Done 关闭 Closer
// 可以通过 返回值 chan 关闭 Closer
func registerCloseCnn0(mgt0 *mgt, c io.Closer) chan bool {
	cc := make(chan bool, 1)

	mgt0.Wg.Add(1)
	go func() {
		select {
		case <-mgt0.WaitCtx.Done():
		case <-cc:
		}
		_ = c.Close()
		mgt0.Wg.Done()
	}()
	return cc
}

// 有的人是双向中任何一个方向断开，都会断开双向
// 代码这样写, 技巧：还使用到了 sync.Once 只运行 1 次。
// var once sync.Once
// go func() {
// 	io.Copy(connection, bashf)
// 	once.Do(close)
// }()
// go func() {
// 	io.Copy(bashf, connection)
// 	once.Do(close)
// }()
func forwardTCP(mgt0 *mgt, c *chain, left io.ReadWriteCloser, right io.ReadWriteCloser) {
	_ = c
	wg := new(sync.WaitGroup)
	canClose := make(chan bool, 1)
	atomic.AddInt64(&mgt0.TcpCnnCnt, 1)

	// right -> left
	mgt0.Wg.Add(1)
	wg.Add(1)
	go func() {
		b := make([]byte, 512*1024)
		_, _ = io.CopyBuffer(left, right, b)
		wg.Done()
		mgt0.Wg.Done()
	}()

	// left -> right
	mgt0.Wg.Add(1)
	wg.Add(1)
	go func() {
		b := make([]byte, 512*1024)
		_, _ = io.CopyBuffer(right, left, b)
		wg.Done()
		mgt0.Wg.Done()
	}()

	// wait read & write close
	mgt0.Wg.Add(1)
	go func() {
		wg.Wait()
		close(canClose)
		mgt0.Wg.Done()
	}()

	select {
	case <-canClose:
	case <-mgt0.WaitCtx.Done():
	}
	_ = left.Close()
	_ = right.Close()
	atomic.AddInt64(&mgt0.TcpCnnCnt, -1)
}

// 只转发 right -> left 方向的 UDP 报文
// SetReadDeadline 协助完成老化功能
func forwardUDP(mgt0 *mgt, us *udpSession) {
	b := make([]byte, 64*1024)
	closeChan := registerCloseCnn0(mgt0, us)
	for {
		t := time.Now().Add(time.Second * time.Duration(mgt0.UdpTTLSec))
		_ = us.ToCnn.SetReadDeadline(t)
		// this is a udp read
		n, err := us.ToCnn.Read(b)
		_ = us.ToCnn.SetReadDeadline(time.Time{})
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				now := time.Now().Unix()
				hit := atomic.LoadInt64(&us.WriteTime)
				// when read timeout, see age or not
				if now > hit && (now-hit) > int64(mgt0.UdpTTLSec) {
					log.Printf("forwardUDP %v aged", us.OwnerChain)
					break
				}
			} else {
				log.Printf("forwardUDP %v Read err= %v", us.OwnerChain, err)
				break
			}
		} else {
			_, err = us.FromCnn.WriteTo(b[:n], us.FromAddr)
			if err != nil {
				log.Printf("forwardUDP %v WriteTo err= %v", us.OwnerChain, err)
				break
			}
		}
	}
	close(closeChan)
	log.Printf("forwardUDP %v exit", us.OwnerChain)
	mgt0.UdpSsns.Delete(us.FromAddr.String())
}

func setupTCPChain(mgt0 *mgt, c *chain) {
	var err error
	log.Printf("setup chain for %v", c)

	var lc net.ListenConfig
	sn, err := lc.Listen(mgt0.WaitCtx, c.Proto, c.ListenAddr)
	if err != nil {
		log.Fatal(err)
	}
	closeChan := registerCloseCnn0(mgt0, sn)
	for {
		cnn, err := sn.Accept()
		if err != nil {
			log.Printf("setupTCPChain %v Accept err= %v", c, err)
			break
		}

		// connect peer
		d := new(net.Dialer)
		toCnn, err := d.DialContext(mgt0.WaitCtx, c.Proto, c.ToAddr)
		if err != nil {
			log.Printf("dial %v %v err= %v", c.Proto, c.ToAddr, err)
			continue
		}
		log.Printf("setupTCPChain got cnn pair %v %v->%v==>%v->%v", c.Proto,
			cnn.RemoteAddr().String(), cnn.LocalAddr().String(),
			toCnn.LocalAddr().String(), toCnn.RemoteAddr().String())
		mgt0.Wg.Add(1)
		go func(arg0 *mgt, arg1 *chain, arg2 io.ReadWriteCloser, arg3 io.ReadWriteCloser) {
			forwardTCP(arg0, arg1, arg2, arg3)
			arg0.Wg.Done()
		}(mgt0, c, cnn, toCnn)

	}
	close(closeChan)
}

func newUdpSsn(mgt0 *mgt, c *chain, fromAddr net.Addr, fromCnn net.PacketConn) *udpSession {
	d := new(net.Dialer)
	toCnn, err := d.DialContext(mgt0.WaitCtx, c.Proto, c.ToAddr)
	if err != nil {
		log.Printf("dail %v err= %v", c.ToAddr, err)
		return nil
	}
	u := &udpSession{}
	u.WriteTime = time.Now().Unix()
	u.ToCnn = toCnn
	u.FromAddr = fromAddr
	u.FromCnn = fromCnn
	u.OwnerChain = c
	return u
}

func setupUDPChain(mgt0 *mgt, c *chain) {
	var err error
	log.Printf("setup chain for %v", c)

	var lc net.ListenConfig
	pktCnn, err := lc.ListenPacket(mgt0.WaitCtx, c.Proto, c.ListenAddr)
	if err != nil {
		log.Fatal(err)
	}
	closeChan := registerCloseCnn0(mgt0, pktCnn)
	rbuf := make([]byte, 128*1024)
	for {
		rsize, raddr, err := pktCnn.ReadFrom(rbuf)
		if err != nil {
			log.Printf("setupUDPChain %v ReadFrom err= %v", c, err)
			break
		}
		var oldUdpSsn *udpSession
		newUdpSsn := newUdpSsn(mgt0, c, raddr, pktCnn)
		if newUdpSsn != nil {
			ssn, loaded := mgt0.UdpSsns.LoadOrStore(raddr.String(), newUdpSsn)
			if loaded {
				_ = newUdpSsn.Close()
			} else {
				log.Printf("setupUDPChain got cnn pair %v %v->%v==>%v->%v", c.Proto,
					raddr.String(), pktCnn.LocalAddr().String(),
					newUdpSsn.ToCnn.LocalAddr().String(), newUdpSsn.ToCnn.RemoteAddr().String())
				mgt0.Wg.Add(1)
				go func(arg0 *mgt, arg1 *udpSession) {
					forwardUDP(arg0, arg1)
					arg0.Wg.Done()
				}(mgt0, newUdpSsn)
			}
			oldUdpSsn, _ = ssn.(*udpSession)
		} else {
			ssn, ok := mgt0.UdpSsns.Load(raddr.String())
			if !ok {
				log.Printf("setupUDPChain %v fail newUdpSsn and fail load from map", c)
			} else {
				oldUdpSsn, _ = ssn.(*udpSession)
			}
		}
		if oldUdpSsn == nil {
			continue
		}
		_, err = oldUdpSsn.ToCnn.Write(rbuf[:rsize])
		atomic.StoreInt64(&oldUdpSsn.WriteTime, time.Now().Unix())
		if err != nil {
			log.Printf("setupUDPChain ToCnn Write err= %v", err)
			_ = oldUdpSsn.Close()
			mgt0.UdpSsns.Delete(raddr.String())
		}
	}
	close(closeChan)
}

/**
放弃的一种配置文件格式
rinetd.toml sample
[[Chans]]
ListenAddr="0.0.0.0:5678"
Proto="tcp"
PeerAddr="127.0.0.1:8100"

[[Chans]]
ListenAddr="0.0.0.0:5679"
Proto="tcp"
PeerAddr="127.0.0.1:8200"

parser sample
0.0.0.0 5678/tcp 127.0.0.1 8100/tcp

用上面的都太复杂了
*/

func setupChains(mgt0 *mgt, cancel context.CancelFunc) {
	setupSignal(mgt0, cancel)
	if len(mgt0.Chains) == 0 {
		log.Printf("no chains to work")
		cancel()
		return
	}
	for _, c := range mgt0.Chains {
		mgt0.Wg.Add(1)
		go func(arg0 *mgt, arg1 *chain) {
			if arg1.Proto == "tcp" {
				setupTCPChain(arg0, arg1)
			} else if arg1.Proto == "udp" {
				setupUDPChain(arg0, arg1)
			}
			arg0.Wg.Done()
		}(mgt0, c)
	}
}

func listChainsFromConf(filename string, mgt0 *mgt) {
	fr, err := os.Open(filename)
	if err != nil {
		log.Fatalf("read %v err= %v", filename, err)
	}

	sc := bufio.NewScanner(fr)
	for sc.Scan() {
		t := sc.Text()
		t = strings.TrimSpace(t)
		if len(t) <= 0 || strings.HasPrefix(t, "#") || strings.HasPrefix(t, "//") {
			continue
		}

		ar := strings.Fields(t)
		if len(ar) < 3 {
			continue
		}
		arValid := make([]string, 0)
		for _, e := range ar {
			e = strings.TrimSpace(e)
			if len(e) > 0 {
				arValid = append(arValid, e)
			}
		}
		if len(arValid) > 2 {
			v := &chain{}
			v.ListenAddr = arValid[0]
			v.ToAddr = arValid[1]
			v.Proto = strings.ToLower(arValid[2])
			mgt0.Chains = append(mgt0.Chains, v)
		}
	}
	_ = fr.Close()
}

func stat(mgt0 *mgt) {
	tc := time.Tick(mgt0.StatInterval)
loop:
	for {
		select {
		case <-mgt0.WaitCtx.Done():
			break loop
		case <-tc:
			log.Printf("tcp %v udp %v", mgt0.TcpCnnCnt, mgt0.UdpCnnCnt())
		}
	}
}

func doWork() {
	var err error
	var cancel context.CancelFunc
	mgt0 := new(mgt)
	mgt0.Wg = new(sync.WaitGroup)
	mgt0.Chains = make([]*chain, 0)
	mgt0.StatInterval = time.Minute
	mgt0.UdpTTLSec = int64(time.Minute.Seconds())
	mgt0.WaitCtx, cancel = context.WithCancel(context.Background())

	fullPath, _ := os.Executable()
	cur := filepath.Dir(fullPath)
	confPath := filepath.Join(cur, "rinetd.conf")

	_, err = os.Stat(confPath)
	if err != nil {
		log.Fatalf("err= %v error of read conf confPath= %v", err, confPath)
	}
	listChainsFromConf(confPath, mgt0)
	// create mgt0.WaitCtx
	setupChains(mgt0, cancel)
	stat(mgt0)
	log.Printf("wait exit")
	mgt0.Wg.Wait()
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix(fmt.Sprintf("pid= %v ", os.Getpid()))
	doWork()
	log.Printf("main exit")
}
