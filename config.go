package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
	"golang.org/x/exp/slog"
)

func parseConfig(r io.Reader) ([]*chain, error) {
	var resultList []*chain

	var scanner = bufio.NewScanner(r)
	for scanner.Scan() {
		var textLine = scanner.Text()
		textLine = strings.TrimSpace(textLine)
		if len(textLine) <= 0 || strings.HasPrefix(textLine, "#") || strings.HasPrefix(textLine, "//") {
			continue
		}

		var lineFieldList = strings.Fields(textLine)
		if len(lineFieldList) < 3 {
			continue
		}
		var validFieldList = make([]string, 0)
		for _, field := range lineFieldList {
			field = strings.TrimSpace(field)
			if len(field) > 0 {
				validFieldList = append(validFieldList, field)
			}
		}
		if len(validFieldList) > 2 {
			v := &chain{}
			v.Proto = strings.ToLower(validFieldList[0])
			v.ListenAddr = validFieldList[1]
			v.ToAddr = validFieldList[2]
			resultList = append(resultList, v)
		}
	}
	return resultList, nil
}

func parseConfigFile(filename string) ([]*chain, error) {
	var fr, err = os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed open file '%s' with error %w", filename, err)
	}
	defer fr.Close()
	return parseConfig(fr)
}

// watchConfig when config file changed, re-parse config file, push result to channel
func watchConfig(filePath string, ch chan []*chain) (func(waitCtx context.Context, logger *slog.Logger), func(), error) {
	if result, err := parseConfigFile(filePath); err != nil {
		return nil, nil, fmt.Errorf("failed parse config file with error %w", err)
	} else {
		ch <- result
	}
	var stringOfEvent = func(e fsnotify.Event) string {
		return fmt.Sprintf("%s: %s", e.Op.String(), e.Name)
	}
	var watcher, err = NewSafeWatcher()
	if err != nil {
		return nil, nil, err
	}
	if err = watcher.Add(filePath); err != nil {
		watcher.Close()
		return nil, nil, err
	}
	watchLoopFunc := func(waitCtx context.Context, logger *slog.Logger) {
		var eventList []fsnotify.Event
		var ok bool
		var result []*chain
	loop:
		for {
			select {
			case eventList, ok = <-watcher.Events:
				for _, e := range eventList {
					logger.Info("got event", "event", stringOfEvent(e))
				}
				if !ok {
					err = fmt.Errorf("closed chan of watcher.Events channel, not receive error")
					logger.Error("exit watch config", "error", err)
					return
				}
				if result, err = parseConfigFile(filePath); err != nil {
					logger.Error("failed parse config file", "error", err)
				} else {
					ch <- result
				}
			case err, ok = <-watcher.Errors:
				if !ok {
					err = fmt.Errorf("closed chan of watcher.Errors channel, not receive error")
					logger.Error("exit watch config", "error", err)
					return
				}
				logger.Error("receive error from watcher.Errors channel", "error", err)
			case <-waitCtx.Done():
				logger.Info("exit watch config for context done")
				break loop
			}
		}
	}
	closeFunc := func() {
		watcher.Close()
	}
	return watchLoopFunc, closeFunc, nil
}
