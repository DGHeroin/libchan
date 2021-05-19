package kcp

import (
    "context"
    "crypto/sha1"
    . "github.com/DGHeroin/libchan"
    "github.com/DGHeroin/libchan/common"
    "github.com/xtaci/kcp-go/v5"
    "golang.org/x/crypto/pbkdf2"
    "net"
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
func (p *kcpTransport) getBlockCrypt() kcp.BlockCrypt {
    u := p.u
    password := u.Query().Get("password")
    salt := u.Query().Get("salt")
    if password == "" || salt == "" {
        return nil
    }
    key := pbkdf2.Key([]byte(password), []byte(salt), 1024, 32, sha1.New)
    block, _ := kcp.NewAESBlockCrypt(key)
    return block
}
func (p *kcpTransport) serve() {
    block := p.getBlockCrypt()
    if block == nil {
        listener, err := kcp.Listen(p.u.Host)
        if err != nil {
            return
        }
        p.closer = listener
        for {
            s, err2 := listener.Accept()
            if err2 != nil {
                break
            }
            go p.handleAcceptSession(s)
        }
    } else {
        listener, err := kcp.ListenWithOptions(p.u.Host, block, 10, 3)
        if err != nil {
            return
        }
        p.closer = listener
        for {
            s, err2 := listener.AcceptKCP()
            if err2 != nil {
                break
            }
            go p.handleAcceptSession(s)
        }
    }
}

func (p *kcpTransport) handleAcceptSession(conn net.Conn) {
    opt := common.ParseConnOption(p.u)
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
func (p *kcpTransport) setupConn(conn net.Conn) {
    cc, ok := conn.(*kcp.UDPSession)
    if cc == nil || !ok {
        return
    }
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
    block := p.getBlockCrypt()
    opt := common.ParseConnOption(p.u)
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
        if conn, err := kcp.DialWithOptions(p.u.Host, block, 10, 3); err == nil {
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
