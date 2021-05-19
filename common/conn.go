package common

import (
    "bytes"
    "context"
    "fmt"
    "net"
    "net/url"
    "sync"
    "time"
)

type (
    Conn struct {
        ctx          context.Context
        conn         net.Conn
        chRecv       chan []byte
        mq           *MQ
        err          error
        opt          *ConnOption
        lastBatching int
        protocol     protocol
        closeOnce sync.Once
    }
    ConnOption struct {
        AutoReconnect bool // 自动重连
        ReadTimeout   time.Duration
        WriteTimeout  time.Duration
        Batching      bool
        BatchingWait  int
        BatchingN     int
        ProtocolType  int
    }
)

func defaultConnOption() *ConnOption {
    return &ConnOption{
        AutoReconnect: true,
        ReadTimeout:   time.Second * 30,
        WriteTimeout:  time.Second * 30,
        Batching:      true,
        BatchingWait:  1000 * 100,
        BatchingN:     1000 * 100,
        ProtocolType:  0,
    }
}

func ParseConnOption(u *url.URL) *ConnOption {
    return &ConnOption{
        AutoReconnect: UrlBool(u, "auto", true),
        ReadTimeout:   UrlDurationSecond(u, "rtime", time.Second*3),
        WriteTimeout:  UrlDurationSecond(u, "wtime", time.Second*3),
        ProtocolType:  UrlInt(u, "protocol", 0),
        Batching:      UrlBool(u, "batching", false),
    }
}

func NewConn(ctx context.Context, conn net.Conn, opt *ConnOption) *Conn {
    if opt == nil {
        opt = defaultConnOption()
    }
    p := &Conn{
        ctx:      ctx,
        conn:     conn,
        chRecv:   make(chan []byte, 100),
        mq:       NewMQ(opt.BatchingWait),
        opt:      opt,
        protocol: newProtocol(opt.ProtocolType),
    }

    go p.startBatchingSend()
    return p
}

func (p *Conn) SetConn(conn net.Conn) {
    p.conn = conn
}

func (p *Conn) Send(data []byte) error {
    if p.opt.Batching {
        return p.sendBatching(data)
    }
    _ = p.conn.SetWriteDeadline(time.Now().Add(p.opt.WriteTimeout))
    _, err := p.protocol.Send(p.conn, data)
    return err
}

func (p *Conn) sendBatching(data []byte) error {
    msg := p.protocol.Pack(data)
    p.mq.Add(msg)
    return nil
}

func (p *Conn) Read() ([]byte, error) {
    pkt, ok := <-p.chRecv
    if !ok {
        // closed
        return nil, fmt.Errorf("closed")
    }
    return pkt, nil
}

func (p *Conn) Context() context.Context {
    return p.ctx
}

func (p *Conn) DoRead() {
    defer func() {
        p.Close()
    }()
    rTimeout := p.opt.ReadTimeout
    conn := p.conn
    for {
        if rTimeout != 0 {
            _ = conn.SetReadDeadline(time.Now().Add(rTimeout))
        }
        msg, err := p.protocol.Read(conn)
        if err != nil {
            break
        }
        p.chRecv <- msg
    }
}

func (p *Conn) startBatchingSend() {
    if !p.opt.Batching {
        return
    }
    var err error
    for err == nil {
        err = p.doBatchingSend()
    }
}
func (p *Conn) doBatchingSend() error {
    arr := p.mq.Wait(time.Millisecond, 1, p.opt.BatchingN)
    if len(arr) == 0 {
        return nil
    }

    bigData := bytes.NewBuffer(nil)
    p.lastBatching = len(arr)
    for _, val := range arr {
        data := val.([]byte)
        bigData.Write(data)
    }
    _, err := p.protocol.SendRaw(p.conn, bigData.Bytes())
    return err
}

func (p *Conn) Close() error {
    var err error
    p.closeOnce.Do(func() {
        err = p.conn.Close()
        close(p.chRecv)
    })
    return err
}

func (p *Conn) Error() error {
    return p.err
}

func (p *Conn) String() string {
    return fmt.Sprintf("batching:%d", p.lastBatching)
}
