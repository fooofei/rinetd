package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	cerrors "github.com/cockroachdb/errors"
	"github.com/fsnotify/fsnotify"
	"github.com/go-logr/logr"
)

func parseConfig(r io.Reader) ([]*chain, error) {
	var result []*chain

	sc := bufio.NewScanner(r)
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
			v.Proto = strings.ToLower(arValid[0])
			v.ListenAddr = arValid[1]
			v.ToAddr = arValid[2]
			result = append(result, v)
		}
	}
	return result, nil
}

func parseConfigFile(filename string) ([]*chain, error) {
	fr, err := os.Open(filename)
	if err != nil {
		return nil, cerrors.CombineErrors(fmt.Errorf("failed open file '%s'", filename), err)
	}
	defer fr.Close()
	return parseConfig(fr)
}

// watchConfig 当配置文件发生变化时，读取配置文件，解析的结果放到 ch 队列中
func watchConfig(waitCtx context.Context, logger logr.Logger, filePath string, ch chan []*chain) (func(), func(), error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, nil, err
	}
	if err = watcher.Add(filePath); err != nil {
		watcher.Close()
		return nil, nil, err
	}
	workFunc := func() {
		var event fsnotify.Event
		var ok bool
		var result []*chain
		var tm = time.NewTimer(time.Second)

	loop:
		for {
			select {
			case <-tm.C: // 当收到 event 时不能直接读取文件，而是通过 timer 读取文件 汇聚连续发生的 event
				if result, err = parseConfigFile(filePath); err != nil {
					logger.Error(err, "failed parse config file", "fileName", event.Name)
				} else {
					ch <- result
				}
			// 一次写会通知两次 需要进行聚合
			// 官方收到很多 issue 但是没有解答
			// https://github.com/fsnotify/fsnotify/issues/122
			// https://github.com/fsnotify/fsnotify/issues/206
			// https://github.com/fsnotify/fsnotify/issues/324
			case event, ok = <-watcher.Events:
				logger.Info("got events", "event", event.String())
				if !ok {
					err = fmt.Errorf("closed chan of watcher.Events, not receive error")
					logger.Error(err, "exit watch config")
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					logger.Info("got file changed", "fileName", event.Name)
					tm.Reset(2 * time.Second)
				} else if event.Op&fsnotify.Create == fsnotify.Create {
					logger.Info("got file create", "fileName", event.Name)
					tm.Reset(2 * time.Second)
				}
			case err, ok = <-watcher.Errors:
				if !ok {
					err = fmt.Errorf("closed chan of watcher.Errors, not receive error")
					logger.Error(err, "exit watch config")
					return
				}
				logger.Error(err, "receive error from watcher.Errors")
			case <-waitCtx.Done():
				logger.Info("exit watch config for context done")
				break loop
			}
		}
	}
	closeFunc := func() {
		watcher.Close()
	}
	return workFunc, closeFunc, nil
}