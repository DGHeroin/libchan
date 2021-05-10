package kcp

import (
    "context"
    "crypto/sha1"
    . "github.com/DGHeroin/libchan"
    "github.com/xtaci/kcp-go/v5"
    "golang.org/x/crypto/pbkdf2"
    "log"
    "net/url"
)

func NewKCP(uri string) Transport {
    k := &kcpTransport{
        uri:    uri,
        chRecv: make(chan []byte, 10),
        chSend: make(chan []byte, 10),
    }
    k.init()
    return k
}

type (
    kcpTransport struct {
        ctx    context.Context
        uri    string
        chSend chan []byte
        chRecv chan []byte
    }
)

func (k *kcpTransport) Context() context.Context {
    return k.ctx
}

func (k *kcpTransport) Send(data []byte) error {
    k.chSend <- data
    return nil
}

func (k *kcpTransport) Recv() ([]byte, error) {
    data, ok := <-k.chRecv
    if !ok {
        // closed
    }
    return data, nil
}

func (k *kcpTransport) init() {
    k.ctx = context.Background()
    u, err := url.Parse(k.uri)
    if err != nil {
        return
    }
    log.Println(">>>", u.Scheme)
    switch u.Scheme {
    case "kcp":
        go k.serve(u)
    case "kcpc":
        go k.client(u)
    }
}

func (k *kcpTransport) serve(u *url.URL) {
    password := u.Query().Get("password")
    salt := u.Query().Get("salt")
    key := pbkdf2.Key([]byte(password), []byte(salt), 1024, 32, sha1.New)
    block, _ := kcp.NewAESBlockCrypt(key)
    if listener, err := kcp.ListenWithOptions(u.Host, block, 10, 3); err == nil {
        log.Println("等待客户端", password, salt)
        for {
            s, err := listener.AcceptKCP()
            if err != nil {
                log.Fatal(err)
            }
            go k.handleAcceptSession(s)
        }
    } else {
        log.Println("server err:", err)
    }
}

func (k *kcpTransport) handleAcceptSession(conn *kcp.UDPSession) {
    buf := make([]byte, 4096)
    for {
        n, err := conn.Read(buf)
        if err != nil {
            log.Println(err)
            return
        }
        k.chRecv <- buf[:n]
    }
}

func (k *kcpTransport) client(u *url.URL) {
    password := u.Query().Get("password")
    salt := u.Query().Get("salt")
    key := pbkdf2.Key([]byte(password), []byte(salt), 1024, 32, sha1.New)
    block, _ := kcp.NewAESBlockCrypt(key)

    if sess, err := kcp.DialWithOptions(u.Host, block, 10, 3); err == nil {
        for {
            select {
            case <-k.ctx.Done():
                return
            case data := <-k.chSend:
                log.Println("发送...", data)
                _, err := sess.Write(data)
                if err != nil {
                    return
                }
            }
        }
    } else {
        log.Println(err)
    }
}
