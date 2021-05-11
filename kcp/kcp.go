package kcp

import (
    "context"
    "crypto/sha1"
    . "github.com/DGHeroin/libchan"
    "github.com/DGHeroin/libchan/common"
    "github.com/xtaci/kcp-go/v5"
    "golang.org/x/crypto/pbkdf2"
    "log"
    "net/url"
    "sync"
)

func New(uri string) Transport {
    k := &kcpTransport{
        uri:      uri,
        acceptCh: make(chan Chan, 10),
    }
    k.init()
    return k
}

type (
    kcpTransport struct {
        ctx      context.Context
        uri      string
        u        *url.URL
        acceptCh chan Chan
        once     sync.Once
        closer   Closer
    }
)

func (p *kcpTransport) Context() context.Context {
    return p.ctx
}

func (p *kcpTransport) init() {
    p.ctx = context.Background()
    u, err := url.Parse(p.uri)
    if err != nil {
        return
    }
    p.u = u
}

func (p *kcpTransport) serve() {
    u := p.u
    password := u.Query().Get("password")
    salt := u.Query().Get("salt")
    key := pbkdf2.Key([]byte(password), []byte(salt), 1024, 32, sha1.New)
    block, _ := kcp.NewAESBlockCrypt(key)
    if listener, err := kcp.ListenWithOptions(u.Host, block, 10, 3); err == nil {
        p.closer = listener
        for {
            s, err := listener.AcceptKCP()
            if err != nil {
                log.Fatal(err)
            }
            go p.handleAcceptSession(s)
        }
    } else {
        log.Println("server err:", err)
    }
}

func (p *kcpTransport) handleAcceptSession(conn *kcp.UDPSession) {
    cli := common.NewConn(p.ctx, conn)
    go func() {
        p.acceptCh <- cli
    }()
    cli.DoRead()
}

func (p *kcpTransport) Accept() Chan {
    p.once.Do(func() {
        go p.serve()
    })
    cli := <-p.acceptCh
    return cli
}

func (p *kcpTransport) Dial() (Chan, error) {
    u := p.u
    password := u.Query().Get("password")
    salt := u.Query().Get("salt")
    key := pbkdf2.Key([]byte(password), []byte(salt), 1024, 32, sha1.New)
    block, _ := kcp.NewAESBlockCrypt(key)
    if conn, err := kcp.DialWithOptions(u.Host, block, 10, 3); err == nil {
        cli := common.NewConn(p.ctx, conn)
        p.closer = cli
        go func() {
            defer conn.Close()
            cli.DoRead()
        }()
        return cli, nil
    } else {
        return nil, err
    }

}
func (p *kcpTransport) Close() error {
    return p.closer.Close()
}
