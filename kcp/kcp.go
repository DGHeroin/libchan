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
    "sync/atomic"
    "time"
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
    u := p.u
    opt := &common.ConnOption{
        AutoReconnect: common.UrlBool(u, "auto", true),
        ReadTimeout: common.UrlDurationSecond(u, "rtime", time.Second*3),
        WriteTimeout: common.UrlDurationSecond(u, "wtime", time.Second*3),
    }
    p.setupConn(conn)

    cli := common.NewConn(p.ctx, conn, opt)
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
func (p *kcpTransport) setupConn(conn *kcp.UDPSession) {
    //普通模式
    //SetNoDelay(32, 32, 0, 40, 0, 0, 100, 1400)
    //极速模式
    //SetNoDelay(32, 32, 1, 10, 2, 1, 30, 1400)

    //conn.SetNoDelay(1, 10, 2, 1)
    //conn.SetACKNoDelay(true)
    //conn.SetStreamMode(true)
    //conn.SetWindowSize(4096, 4096)
    //_ = conn.SetReadBuffer(4096)
}

func (p *kcpTransport) Dial() (Chan, error) {
    u := p.u
    password := common.UrlString(u, "password", "")
    salt := common.UrlString(u, "salt", "")
    key := pbkdf2.Key([]byte(password), []byte(salt), 1024, 32, sha1.New)
    block, _ := kcp.NewAESBlockCrypt(key)

    opt := &common.ConnOption{
        AutoReconnect: common.UrlBool(u, "auto", true),
        ReadTimeout: common.UrlDurationSecond(u, "rtime", time.Second*3),
        WriteTimeout: common.UrlDurationSecond(u, "wtime", time.Second*3),
    }
    cli := common.NewConn(p.ctx, nil, opt)
    var (
        dial        func() error
        isDialing   int32
        isConnected = false
    )

    dial = func() error {
        if atomic.LoadInt32(&isDialing) == 1 {
            return nil
        }
        if conn, err := kcp.DialWithOptions(u.Host, block, 10, 3); err == nil {
            p.closer = cli
            cli.SetConn(conn)
            p.setupConn(conn)
            go func() {
                isConnected = true
                defer func() {
                    conn.Close()
                    if cli.Error() != nil && opt.AutoReconnect {
                        time.AfterFunc(time.Second*5, func() {
                            dial()
                        })
                    }
                }()
                cli.DoRead()
            }()
            return nil
        } else {
            atomic.StoreInt32(&isDialing, 0)
            if isConnected == true && opt.AutoReconnect { // 曾经链接成功
                time.AfterFunc(time.Second*5, func() {
                    dial()
                })
            }
            return err
        }
    }
    return cli, dial()
}
func (p *kcpTransport) Close() error {
    return p.closer.Close()
}
