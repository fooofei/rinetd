package main

import (
	"context"
	"io"
	stdlog "log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
)

// worked like `rinetd`.

// 可以通过 WaitCtx.Done 关闭 Closer
// 可以通过 返回值 chan 关闭 Closer
func closeWhenContext(waitCtx context.Context, c io.Closer) chan struct{} {
	cc := make(chan struct{}, 1)
	go func() {
		select {
		case <-waitCtx.Done():
		case <-cc:
		}
		c.Close()
	}()
	return cc
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
tcp 0.0.0.0:5678 127.0.0.1:8100

用上面的都太复杂了
*/

// reconcileListeners 将会调协
// curChains 是当前使用的 chain
// expectChains 是从配置文件读取的最新的 chain
// 返回合并过的 chain
func reconcileListeners(waitCtx context.Context, logger logr.Logger, wg *sync.WaitGroup,
	curChains []*chain, expectChains []*chain) []*chain {
	var createdCnt int64
	var keepedCnt int64
	var closedCnt int64

	chains := make(map[string]*chain, len(curChains))
	for _, c := range curChains {
		chains[c.String()] = c
	}
	var mergedChains []*chain
	for _, c := range expectChains {
		findChain, ok := chains[c.String()]
		if !ok {
			createChainRoutine(waitCtx, logger, wg, c)
			mergedChains = append(mergedChains, c)
			createdCnt += 1
		} else {
			mergedChains = append(mergedChains, findChain)
			delete(chains, c.String())
			keepedCnt += 1
		}
	}

	// 老的老化掉
	for _, c := range chains {
		c.Cancel()
		closedCnt += 1
	}
	logger.Info("reconcile listeners",
		"createdChainCnt", createdCnt, "keepedChainCnt", keepedCnt, "closedChainCnt", closedCnt,
		"beforeChainCount", len(curChains), "afterChainCount", len(mergedChains))
	return mergedChains
}

// doWork 工作循环，不断的监视文件改动 根据文件实际内容进行 chain 调谐
func doWork(waitCtx context.Context, logger logr.Logger, filePath string) {
	var chains []*chain
	var statInterval = 3 * time.Second
	var chainsCh = make(chan []*chain, 10)
	var expectChains []*chain
	var wg = &sync.WaitGroup{}

	tc := time.NewTicker(statInterval)
	defer tc.Stop()

	cfgWrk, cfgClose, err := watchConfig(waitCtx, logger, filePath, chainsCh)
	if err != nil {
		logger.Error(err, "failed create watch config")
		return
	}
	defer cfgClose()
	go cfgWrk()

loop:
	for {
		select {
		case <-tc.C:
			summary(logger, chains)
		case expectChains = <-chainsCh:
			chains = reconcileListeners(waitCtx, logger, wg, chains, expectChains)
		case <-waitCtx.Done():
			break loop
		}
	}
	wg.Wait()
	logger.Info("exit of do work")
}

func main() {
	// setup logger in main routine
	logger := stdr.New(stdlog.New(os.Stderr, "", stdlog.Lshortfile|stdlog.LstdFlags))
	logger = logger.WithValues("pid", os.Getpid())
	//
	var err error
	fullPath, _ := os.Executable()
	cur := filepath.Dir(fullPath)
	confPath := filepath.Join(cur, "rinetd.conf")
	_, err = os.Stat(confPath)
	if err != nil {
		logger.Error(err, "error config file, not exists", "filepath", confPath)
		return
	}
	//
	waitCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	doWork(waitCtx, logger, confPath)
	logger.Info("main exit")
}
