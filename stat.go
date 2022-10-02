package main

import (
	"sync/atomic"

	"github.com/go-logr/logr"
)

func summary(logger logr.Logger, chains []*chain) {
	var tcpCnt int64
	var udpCnt int64

	for _, c := range chains {
		if c.Proto == "tcp" {
			tcpCnt += atomic.LoadInt64(&c.SsnCount)
		} else if c.Proto == "udp" {
			udpCnt += atomic.LoadInt64(&c.SsnCount)
		}
	}
	logger.V(9).Info("stat count", "tcpSession", tcpCnt, "udpSession", udpCnt, "chainCount", len(chains))
}
