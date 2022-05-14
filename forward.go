package main

import (
	"context"
	"io"
	"sync"
)

// readWriteEach bind two ReadWriter which we called left and right
// when both cannot left -> right and right -> left we exit this function
//
// someone will write:
// var once sync.Once
// go func() {
// 	io.Copy(connection, bashf)
// 	once.Do(close)
// }()
// go func() {
// 	io.Copy(bashf, connection)
// 	once.Do(close)
// }()
func readWriteEach(waitCtx context.Context, left io.ReadWriteCloser, right io.ReadWriteCloser) {
	wg := &sync.WaitGroup{}
	bothDone := make(chan struct{})
	// right -> left
	wg.Add(1)
	go func() {
		b := make([]byte, 512*1024)
		io.CopyBuffer(left, right, b)
		wg.Done()
	}()

	// left -> right
	wg.Add(1)
	go func() {
		b := make([]byte, 512*1024)
		io.CopyBuffer(right, left, b)
		wg.Done()
	}()
	// wait read & write close
	go func() {
		wg.Wait()
		close(bothDone)
	}()
	select {
	case <-bothDone:
	case <-waitCtx.Done():
	}
	left.Close()
	right.Close()
	wg.Wait()
}
