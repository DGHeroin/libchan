package common

import (
    "bytes"
    "container/list"
    "context"
    "encoding/binary"
    "io"
    "net"
    "sync"
    "time"
)

type (
    Conn struct {
        ctx      context.Context
        chRecv   chan []byte
        sendList list.List
        mu       sync.Mutex
    }
)
func NewConn(ctx context.Context, conn net.Conn) *Conn {
    ctx2 := context.WithValue(ctx, "conn", conn)
    p := &Conn{
        ctx:    ctx2,
        chRecv: make(chan []byte, 100),
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
    p.mu.Lock()
    defer p.mu.Unlock()

    header := make([]byte, 4)
    binary.BigEndian.PutUint32(header, uint32(len(data)))
    msg := append(header, data...)
    p.sendList.PushBack(msg)

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
        time.Sleep(time.Millisecond * 10)
        p.doBatchingSend()
    }
}
func (p *Conn) doBatchingSend() {
    p.mu.Lock()
    defer p.mu.Unlock()
   // startTime := time.Now()
    conn := p.ctx.Value("conn").(net.Conn)

    sz := p.sendList.Len()
    if sz == 0 {
        return
    }

    bigData := bytes.NewBuffer(nil)
    it := p.sendList.Front()
    n := 0
    for it != nil {
        data := it.Value.([]byte)
        bigData.Write(data)
        it = it.Next()
        n++
    }
    p.sendList.Init()

    conn.Write(bigData.Bytes())

//    log.Println(bigData.Len(),time.Now().Sub(startTime))
}
