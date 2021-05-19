package common

import (
    "net/url"
    "time"
)

type (
    ConnOption struct {
        AutoReconnect bool // 自动重连
        ReadTimeout   time.Duration
        WriteTimeout  time.Duration
        Batching      bool
        BatchingWait  int
        BatchingN     int
        ProtocolType  int
        Compression   int //0 none, 1 gzip
    }
)

func ParseConnOption(u *url.URL) *ConnOption {
    return &ConnOption{
        AutoReconnect: UrlBool(u, "auto", true),
        ReadTimeout:   UrlDurationSecond(u, "rtime", time.Second*3),
        WriteTimeout:  UrlDurationSecond(u, "wtime", time.Second*3),
        ProtocolType:  UrlInt(u, "protocol", 0),
        Batching:      UrlBool(u, "batching", false),
        Compression:   UrlInt(u, "compression", 0),
    }
}
func defaultConnOption() *ConnOption {
    return &ConnOption{
        AutoReconnect: true,
        ReadTimeout:   time.Second * 30,
        WriteTimeout:  time.Second * 30,
        Batching:      true,
        BatchingWait:  1000 * 100,
        BatchingN:     1000 * 100,
        ProtocolType:  0,
        Compression:   0,
    }
}
