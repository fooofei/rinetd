package main

import (
    "context"
    "fmt"
    "io"
    "log"
    "net"
    "os"
    "os/signal"
    "path/filepath"
    "sync"
    "syscall"
)

//
//
// work like rinetd.


type cchan struct {
    ListenAddr string `toml:"ListenAddr"`
    Proto      string `toml:"Proto"`
    PeerAddr   string `toml:"PeerAddr"`
}

func (c *cchan) String() string {
    return fmt.Sprintf("ListenAddr=%v Proto=%v PeerAddr=%v",
        c.ListenAddr, c.Proto, c.PeerAddr)
}

type mgt struct {
    Chans   []*cchan `toml:"Chans"`
    WaitCtx context.Context
    Wg      *sync.WaitGroup
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

func forward(mgt0 *mgt, c *cchan, left io.ReadWriteCloser, right io.ReadWriteCloser) {
    _ = c
    wg := new(sync.WaitGroup)
    canClose := make(chan bool, 1)

    mgt0.Wg.Add(1)
    wg.Add(1)
    go func() {
        b := make([]byte, 512*1024)
        _, _ = io.CopyBuffer(left, right, b)
        wg.Done()
        mgt0.Wg.Done()
    }()

    mgt0.Wg.Add(1)
    wg.Add(1)
    go func() {
        b := make([]byte, 512*1024)
        _, _ = io.CopyBuffer(right, left, b)
        wg.Done()
        mgt0.Wg.Done()
    }()

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

}

func setupCChanTCP(mgt0 *mgt, c *cchan) {
    var err error
    log.Printf("setup cchan for %v", c)

    var lc net.ListenConfig
    sn, err := lc.Listen(mgt0.WaitCtx, c.Proto, c.ListenAddr)
    if err != nil {
        log.Fatal(err)
    }
    closeChan := registerCloseCnn0(mgt0, sn)
    for {
        cnn, err := sn.Accept()
        if err != nil {
            break
        }

        // connect peer
        d := new(net.Dialer)
        peerCnn, err := d.DialContext(mgt0.WaitCtx, c.Proto, c.PeerAddr)
        if err != nil {
            log.Printf("dial %v err=%v", c.PeerAddr, err)
            _ = cnn.Close()
        } else {
            mgt0.Wg.Add(1)
            go func(arg0 *mgt, arg1 *cchan, arg2 io.ReadWriteCloser, arg3 io.ReadWriteCloser) {
                forward(arg0, arg1, arg2, arg3)
                arg0.Wg.Done()
            }(mgt0, c, cnn, peerCnn)
        }

    }

    close(closeChan)

}

/**
rinetd.toml sample
[[Chans]]
ListenAddr="0.0.0.0:5678"
Proto="tcp"
PeerAddr="127.0.0.1:8100"

[[Chans]]
ListenAddr="0.0.0.0:5679"
Proto="tcp"
PeerAddr="127.0.0.1:8200"
*/

func main1(mgt0 * mgt){
    var cancel context.CancelFunc
    mgt0.WaitCtx, cancel = context.WithCancel(context.Background())
    setupSignal(mgt0, cancel)

    if len(mgt0.Chans) == 0 {
        log.Printf("no chans to work")
        cancel()
    } else {
        for _, c := range mgt0.Chans {
            mgt0.Wg.Add(1)
            go func(arg0 *mgt, arg1 *cchan) {
                setupCChanTCP(arg0, arg1)
                arg0.Wg.Done()
            }(mgt0, c)
        }
        // infinit wait
        <-mgt0.WaitCtx.Done()
    }
}

func main() {

    var err error
    mgt0 := new(mgt)
    mgt0.Wg = new(sync.WaitGroup)

    //
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    log.SetPrefix(fmt.Sprintf("pid= %v ", os.Getpid()))

    fullPath, _ := os.Executable()
    cur := filepath.Dir(fullPath)
    confPath := filepath.Join(cur, "rinetd.conf")

    _,err = os.Stat(confPath)
    if err != nil {
        //log.Fatal(err)
        confPath = "/Users/hujianfei/Desktop/git_src/go_pieces/rinetd/rinetd.conf"
    }

    r,err := ParseFile(confPath, Debug(false), Recover(false))

    if err != nil {
        log.Printf("parser err=%v", err)
    }

    if ar,ok := r.([]*Unit) ;ok {
        log.Printf("ar len=%v\n", len(ar))
        for _,a := range ar {
            log.Println(a)
        }
    }else{
        log.Printf("fail parse, got %v\n",r)
    }

    log.Printf("wait exit")
    mgt0.Wg.Wait()
    log.Printf("main exit")

}
