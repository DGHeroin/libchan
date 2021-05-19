package tcp

import (
    "context"
    . "github.com/DGHeroin/libchan"
    "github.com/DGHeroin/libchan/common"
    "net"
    "net/url"
    "sync"
)

func New(uri string) Transport {
    k := &tcpTransport{
        uri:      uri,
        acceptCh: make(chan Chan, 10),
    }
    k.init()
    return k
}

type (
    tcpTransport struct {
        ctx      context.Context
        uri      string
        u        *url.URL
        acceptCh chan Chan
        once     sync.Once
        closer   Closer
    }
)

func (p *tcpTransport) Context() context.Context {
    return p.ctx
}

func (p *tcpTransport) init() {
    p.ctx = context.Background()
    u, err := url.Parse(p.uri)
    if err != nil {
        return
    }
    p.u = u
}

func (p *tcpTransport) serve() {
    u := p.u
    if listener, err := net.Listen("tcp", u.Host); err == nil {
        for {
            s, err2 := listener.Accept()
            if err2 != nil {
                break
            }
            go p.handleAcceptSession(s)
        }
    }
}

func (p *tcpTransport) handleAcceptSession(conn net.Conn) {
    opt := common.ParseConnOption(p.u)
    cli := common.NewConn(p.ctx, conn, opt)
    go func() {
        p.acceptCh <- cli
    }()
    cli.DoRead()
}

func (p *tcpTransport) Accept() Chan {
    p.once.Do(func() {
        go p.serve()
    })
    cli := <-p.acceptCh
    return cli
}

func (p *tcpTransport) Dial() (Chan, error) {
    u := p.u
    if conn, err := net.Dial("tcp", u.Host); err == nil {
        opt := common.ParseConnOption(p.u)
        cli := common.NewConn(p.ctx, conn, opt)
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
func (p *tcpTransport) Close() error {
    return p.closer.Close()
}
