package common

import (
    "bytes"
    "context"
    "encoding/binary"
    "io"
    "net"
    "time"
)

type (
    Conn struct {
        ctx      context.Context
        chRecv   chan []byte
        //sendList list.List
        //mu       sync.Mutex
        mq *MQ
    }
)
func NewConn(ctx context.Context, conn net.Conn) *Conn {
    ctx2 := context.WithValue(ctx, "conn", conn)
    p := &Conn{
        ctx:    ctx2,
        chRecv: make(chan []byte, 100),
        mq: NewMQ(),
    }
    //p.cond = sync.NewCond(&p.mu)
    go p.startBatchingSend()
    return p
}
func (p *Conn) Send(data []byte) error {
    conn := p.ctx.Value("conn").(net.Conn)
    header := make([]byte, 4)
    binary.BigEndian.PutUint32(header, uint32(len(data)))
    msg := append(header, data...)
    _, err := conn.Write(msg)
    return err
}

func (p *Conn) SendBatching(data []byte) error {
    header := make([]byte, 4)
    binary.BigEndian.PutUint32(header, uint32(len(data)))
    msg := append(header, data...)
    p.mq.Add(msg)

    return nil
}

func (p *Conn) Read() ([]byte, error) {
    pkt, ok := <-p.chRecv
    if !ok {
        // closed
    }
    return pkt, nil
}

func (p *Conn) Context() context.Context {
    return p.ctx
}

func (p *Conn) DoRead() {
    conn := p.ctx.Value("conn").(net.Conn)
    for {
        header := make([]byte, 4)
        n, err := io.ReadFull(conn, header)
        if err != nil {
            return
        }
        size := binary.BigEndian.Uint32(header)
        body := make([]byte, size)
        n, err = io.ReadFull(conn, body)
        if err != nil {
            return
        }
        p.chRecv <- body[:n]
    }
}

func (p *Conn) startBatchingSend() {
    for {
        p.doBatchingSend()
    }
}
func (p *Conn) doBatchingSend() {
    conn := p.ctx.Value("conn").(net.Conn)
    arr := p.mq.Wait(time.Millisecond, 10, 10)
    if len(arr) == 0 {
        return
    }

    bigData := bytes.NewBuffer(nil)
    for _, val := range arr {
        data := val.([]byte)
        bigData.Write(data)
    }
    conn.Write(bigData.Bytes())
}

func (p*Conn) Close() error {
    conn := p.ctx.Value("conn").(net.Conn)
    return conn.Close()
}