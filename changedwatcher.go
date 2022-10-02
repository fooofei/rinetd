package main

import (
	"context"
	"github.com/fooofei/timer/go/timer"
	"github.com/fsnotify/fsnotify"
	"time"
)

// ChangedWatcher is safely watch a file changed or not,
// 不同点：延迟通知文件改变（延迟2s）、汇聚快速变化的多个事件
// fix fsnotify.Watcher 's bug. fsnotify.Watcher will not watch when file is changed with got multi events
//
//	file changed once we will got twice events
//	    issues:
//	    https://github.com/fsnotify/fsnotify/issues/122
//	    https://github.com/fsnotify/fsnotify/issues/206
//	    https://github.com/fsnotify/fsnotify/issues/324
type ChangedWatcher struct {
	fsWatcher *fsnotify.Watcher
	Events    chan []fsnotify.Event
	Errors    chan error
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewSafeWatcher create a watcher
func NewSafeWatcher() (*ChangedWatcher, error) {
	var w, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	var ctx, cancel = context.WithCancel(context.Background())
	var safeWatcher = &ChangedWatcher{
		fsWatcher: w,
		Events:    make(chan []fsnotify.Event, 50),
		Errors:    make(chan error, 10),
		ctx:       ctx,
		cancel:    cancel,
	}
	go safeWatcher.readEvents()
	return safeWatcher, nil
}

// Close removes all watches and closes the events channel.
func (w *ChangedWatcher) Close() error {
	w.cancel()
	return w.fsWatcher.Close()
}

// Add starts watching the named file or directory (non-recursively).
func (w *ChangedWatcher) Add(name string) error {
	return w.fsWatcher.Add(name)
}

func (w *ChangedWatcher) readEvents() {
	var tm = timer.New(time.Second)
	defer tm.Stop()
	var eventList []fsnotify.Event
	tm.Stop()
loop:
	for {
		var event fsnotify.Event
		var err error
		var ok bool
		select {
		case <-w.ctx.Done():
			break loop
		case <-tm.Wait():
			tm.SetUnActive()
			if len(eventList) < 1 {
				continue
			}
			select {
			case w.Events <- eventList:
			case <-w.ctx.Done():
			}
			eventList = eventList[0:]
		case event, ok = <-w.fsWatcher.Events:
			if !ok {
				close(w.Events)
				break loop
			}
			// when file is removed, we force watch it back
			// this occurs when edit file with `vi xx.conf`, we will got RENAME + CHMOD + REMOVE events
			if event.Op&fsnotify.Remove == fsnotify.Remove {
				w.fsWatcher.Add(event.Name)
			}
			// all event will be set timer for read
			// because we cannot trust the events
			// use vi modify once, we will receive three events RENAME+CHMOD+REMOVE， it will not trigger read file again
			tm.Reset(2 * time.Second)
			eventList = append(eventList, event)
		case err, ok = <-w.fsWatcher.Errors:
			if !ok {
				close(w.Errors)
				break loop
			}
			select {
			case w.Errors <- err:
			case <-w.ctx.Done():
			}
		}
	}
}
