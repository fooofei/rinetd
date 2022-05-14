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

// closeWhenContext will close c when context done or close returned channel
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

// reconcileListeners will make listens as config file expected
// it will close deleted listeners, stay keeped listeners open, create new listeners
// arg curChains is current chains
// arg expectChains is the expect chains read from config file
// return the final chains
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

	// close delete chains
	for _, c := range chains {
		c.Cancel()
		closedCnt += 1
	}
	logger.Info("reconcile listeners",
		"createdChainCnt", createdCnt, "keepedChainCnt", keepedCnt, "closedChainCnt", closedCnt,
		"beforeChainCount", len(curChains), "afterChainCount", len(mergedChains))
	return mergedChains
}

// doWork loop in stat summary + watch config file + reconcile listeners
func doWork(waitCtx context.Context, logger logr.Logger, filePath string) {
	var chains []*chain
	var statInterval = time.Minute
	var chainsCh = make(chan []*chain, 10)
	var expectChains []*chain
	var wg = &sync.WaitGroup{}

	tc := time.NewTicker(statInterval)
	defer tc.Stop()

	cfgWrk, cfgClose, err := watchConfig(filePath, chainsCh)
	if err != nil {
		logger.Error(err, "failed create watch config")
		return
	}
	defer cfgClose()
	go cfgWrk(waitCtx, logger)

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
		logger.Error(err, "file stat config file", "filePath", confPath)
		return
	}
	//
	waitCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	doWork(waitCtx, logger, confPath)
	logger.Info("main exit")
}
