package main

import (
    "github.com/DGHeroin/libchan"
    "github.com/DGHeroin/libchan/kcp"
    "log"
)

func server() {
    ch := kcp.New("tcp://127.0.0.1:6000?password=aoe&salt=123")
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
    server()
}
