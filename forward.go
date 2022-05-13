package main

import (
	"context"
	"io"
	"sync"
)

// readWriteEach 把两个 ReadWriter 互相连接起来，进行转发
// 当双向都无法读写时 函数才退出
//
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
