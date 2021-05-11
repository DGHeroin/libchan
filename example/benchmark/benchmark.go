package main

import (
    "github.com/DGHeroin/libchan"
    "github.com/DGHeroin/libchan/transport"
    "github.com/DGHeroin/libchan/common"
    "log"
    "sync/atomic"
    "time"
)

var (
    sendQPS     uint32
    recvQPS     uint32
    pktQPS      uint32
    bandwidth   uint32
    latestBytes []byte
    rawurl = "kcp://127.0.0.1:6000?password=aoe&salt=123"
)

func client() {
    time.Sleep(time.Second)
    cc := transport.New(rawurl)
    ch, err := cc.Dial()
    if err != nil {

    }
    go func() {
        for {
            time.Sleep(time.Second)
            s := atomic.SwapUint32(&sendQPS, 0)
            r := atomic.SwapUint32(&recvQPS, 0)
            p := atomic.SwapUint32(&pktQPS, 0)
            b := atomic.SwapUint32(&bandwidth, 0)
            log.Println("qps", s, r, p, "==>", string(latestBytes), common.ByteSize(uint64(b)))
        }
    }()
    sendData := make([]byte, 1000)
    go func() {
        for {
            ch.SendBatching(sendData)
            atomic.AddUint32(&sendQPS, 1)
        }
    }()
    for {
        _, err := ch.Read()
        if err != nil {
            return
        }
        atomic.AddUint32(&recvQPS, 1)
    }
}
func server() {
    ch := transport.New(rawurl)
    for {
        remote := ch.Accept()
        go func(remote libchan.Chan) {
            for {
                data, err := remote.Read()
                if err != nil {
                    return
                }
                atomic.AddUint32(&pktQPS, 1)
                latestBytes = data
                atomic.AddUint32(&bandwidth, uint32(len(data)))
                //remote.Send(data)
            }
        }(remote)
    }
}

func main() {
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    go server()
    client()
}
