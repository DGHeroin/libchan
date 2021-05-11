package main

import (
    "fmt"
    "github.com/DGHeroin/libchan"
    "github.com/DGHeroin/libchan/kcp"
    "log"
    "time"
)

func client() {
    cc := kcp.New("kcpc://127.0.0.1:6000?password=aoe&salt=123")
    ch, err := cc.Dial()
    if err != nil {

    }
    go func() {
        for {
            time.Sleep(time.Second)
            ch.Send([]byte(fmt.Sprintf("Hello:%v", time.Now())))
        }
    }()
    for {
        data, err := ch.Read()
        if err != nil {
            return
        }

        log.Println("client recv:", string(data))
    }
}
func server() {
    ch := kcp.New("kcp://127.0.0.1:6000?password=aoe&salt=123")
    for {
        remote := ch.Accept()
        go func(remote libchan.Chan) {
            for {
                data, err := remote.Read()
                if err != nil {
                    return
                }
                log.Println("server recv:", string(data))
                remote.Send(data)
            }
        }(remote)
    }
}

func main() {
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    go server()
    client()
}
