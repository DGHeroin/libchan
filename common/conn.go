package common

import (
    "bytes"
    "context"
    "encoding/binary"
    "fmt"
    "io"
    "log"
    "net"
    "time"
)

type (
    Conn struct {
        ctx    context.Context
        conn   net.Conn
        chRecv chan []byte
        mq     *MQ
        err    error
        opt    *ConnOption
    }
    ConnOption struct {
        AutoReconnect bool // 自动重连
        ReadTimeout   time.Duration
        WriteTimeout  time.Duration
        Batching      bool
    }
)

func defaultConnOption() *ConnOption {
    return &ConnOption{
        AutoReconnect: true,
        ReadTimeout:   time.Second * 30,
        WriteTimeout:  time.Second * 30,
        Batching:      true,
    }
}

func NewConn(ctx context.Context, conn net.Conn, opt *ConnOption) *Conn {
    if opt == nil {
        opt = defaultConnOption()
    }
    p := &Conn{
        ctx:    ctx,
        conn:   conn,
        chRecv: make(chan []byte, 100),
        mq:     NewMQ(),
        opt:    opt,
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
    header := make([]byte, 5)
    header[4] = 0x98
    binary.BigEndian.PutUint32(header, uint32(len(data)))
    msg := append(header, data...)
    _ = p.conn.SetWriteDeadline(time.Now().Add(p.opt.WriteTimeout))
    _, err := p.conn.Write(msg)
    return err
}

func (p *Conn) sendBatching(data []byte) error {
    header := make([]byte, 5)
    header[4] = 0x98
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
    for {
        p.conn.SetReadDeadline(time.Now().Add(p.opt.ReadTimeout))
        header := make([]byte, 5)
        n, err := io.ReadFull(p.conn, header)
        if err != nil {
            p.err = err
            return
        }
        msgT := header[4]
        switch msgT {
        case 0x98:
            size := binary.BigEndian.Uint32(header)
            body := make([]byte, size)
            n, err = io.ReadFull(p.conn, body)
            if err != nil {
                p.err = err
                return
            }
            p.chRecv <- body[:n]
        default:
            log.Println(msgT)
            p.err = fmt.Errorf("unsupport msg type")
        }

    }
}

func (p *Conn) startBatchingSend() {
    for {
        p.doBatchingSend()
    }
}
func (p *Conn) doBatchingSend() {
    arr := p.mq.Wait(time.Millisecond, 10, 10)
    if len(arr) == 0 {
        return
    }

    bigData := bytes.NewBuffer(nil)
    for _, val := range arr {
        data := val.([]byte)
        bigData.Write(data)
    }
    _, err := p.conn.Write(bigData.Bytes())
    if err != nil {
        p.err = err
    }
}

func (p *Conn) Close() error {
    conn := p.ctx.Value("conn").(net.Conn)
    return conn.Close()
}

func (p *Conn) Error() error {
    return p.err
}
