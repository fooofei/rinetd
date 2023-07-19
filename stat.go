package main

import (
	"sync/atomic"

	"golang.org/x/exp/slog"
)

func summary(logger *slog.Logger, chains []*chain) {
	var tcpCnt int64
	var udpCnt int64

	for _, c := range chains {
		if c.Proto == "tcp" {
			tcpCnt += atomic.LoadInt64(&c.SsnCount)
		} else if c.Proto == "udp" {
			udpCnt += atomic.LoadInt64(&c.SsnCount)
		}
	}
	logger.Info("stat count", "tcpSession", tcpCnt, "udpSession", udpCnt, "chainCount", len(chains))
}
