package main

import (
    "fmt"
    "github.com/DGHeroin/libchan/kcp"
    "log"
    "time"
)

func client() {
    cc := kcp.New("tcp://127.0.0.1:6000?password=aoe&salt=123&auto=true")
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

func main()  {
    client()
}