package main

import (
    "fmt"
    "github.com/DGHeroin/libchan/kcp"
    "log"
    "time"
)

func client() {
    ch := kcp.NewKCP("kcpc://127.0.0.1:6000?password=aoe&salt=123")

    for {
        time.Sleep(time.Second)
        ch.Send([]byte(fmt.Sprintf("Hello:%v", time.Now())))
    }
}
func server() {
    ch := kcp.NewKCP("kcp://127.0.0.1:6000?password=aoe&salt=123")
    go func() {
        for {
            time.Sleep(time.Second)
            ch.Send([]byte(fmt.Sprintf("Hello:%v", time.Now())))
        }
    }()

    for {
        data, err := ch.Recv()
        if err != nil {
            return
        }
        log.Println("recv:", string(data))
    }
}

func main() {
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    go server()
    client()
}
