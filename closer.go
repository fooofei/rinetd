package main

import (
	"context"
	"io"
)

// closeWhenContext will close c when context done or close returned channel
func closeWhenContext(waitCtx context.Context, c io.Closer) chan struct{} {
	var cc = make(chan struct{}, 1)
	go func() {
		select {
		case <-waitCtx.Done():
		case <-cc:
		}
		c.Close()
	}()
	return cc
}
