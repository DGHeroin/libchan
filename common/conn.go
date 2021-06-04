package common

import (
    "bytes"
    "context"
    "fmt"
    "net"
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
        closeOnce    sync.Once
    }
)

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

func (p *Conn) Write(data []byte) (int, error) {
    if p.opt.Batching {
        return len(data), p.sendBatching(data)
    }
    data = doCompression(p.opt.Compression, data)
    _ = p.conn.SetWriteDeadline(time.Now().Add(p.opt.WriteTimeout))
    return p.protocol.Send(p.conn, data)
}

func (p *Conn) sendBatching(data []byte) error {
    data = doCompression(p.opt.Compression, data)
    msg := p.protocol.Pack(data)
    p.mq.Add(msg)
    return nil
}
func (p *Conn) Read(data []byte) (int, error) {
    pkt, err := p.ReadMessage()
    if err != nil {
        return -1, err
    }
    sz := len(pkt)
    copy(data, pkt)
    return sz, nil
}
func (p *Conn) ReadMessage() (  []byte, error) {
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
        msg = doUnCompression(p.opt.Compression, msg)
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
    return fmt.Sprintf("[%p] bat:%d", p, p.lastBatching)
}
