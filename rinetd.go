package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"golang.org/x/exp/slog"
)

// worked like `rinetd` https://github.com/samhocevar/rinetd.

// reconcileListeners will make listens as config file expected
// it will close deleted listeners, stay keeped listeners open, create new listeners
// arg curChains is current chains
// arg expectChains is the expect chains read from config file
// return the final chains
func reconcileListeners(waitCtx context.Context, logger *slog.Logger, wg *sync.WaitGroup,
	curChains []*chain, expectChains []*chain) []*chain {
	var createdCnt int64
	var keepedCnt int64
	var closedCnt int64

	var chainMap = make(map[string]*chain, len(curChains))
	for _, c := range curChains {
		chainMap[c.String()] = c
	}
	var mergedChains []*chain
	for _, c := range expectChains {
		findChain, ok := chainMap[c.String()]
		if !ok {
			createChainRoutine(waitCtx, logger, wg, c)
			mergedChains = append(mergedChains, c)
			createdCnt += 1
		} else {
			mergedChains = append(mergedChains, findChain)
			delete(chainMap, c.String())
			keepedCnt += 1
		}
	}

	// close delete chainMap
	for _, c := range chainMap {
		c.Cancel()
		closedCnt += 1
	}
	logger.Info("reconcile listeners",
		"createdChainCnt", createdCnt, "keepedChainCnt", keepedCnt, "closedChainCnt", closedCnt,
		"beforeChainCount", len(curChains), "afterChainCount", len(mergedChains))
	return mergedChains
}

// workLoop loop in stat summary + watch config file + reconcile listeners
func workLoop(waitCtx context.Context, logger *slog.Logger, filePath string) {
	var chains []*chain
	var statInterval = time.Minute
	var chainsCh = make(chan []*chain, 10)
	var expectChains []*chain
	var wg = &sync.WaitGroup{}

	var tc = time.NewTicker(statInterval)
	defer tc.Stop()

	var watchConfigLoopFunc, closeConfigWatch, err = watchConfig(filePath, chainsCh)
	if err != nil {
		logger.Error("failed create watch config", "error", err)
		return
	}
	defer closeConfigWatch()
	go watchConfigLoopFunc(waitCtx, logger)

	// start stat
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
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{}))
	logger = logger.With("pid", os.Getpid())
	//
	var err error
	fullPath, _ := os.Executable()
	cur := filepath.Dir(fullPath)
	confPath := filepath.Join(cur, "rinetd.conf")
	_, err = os.Stat(confPath)
	if err != nil {
		logger.Error("file stat config file", "filePath", confPath, "error", err)
		return
	}
	//
	waitCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	workLoop(waitCtx, logger, confPath)
	logger.Info("main exit")
}
