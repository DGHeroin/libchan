package common

import (
    "encoding/binary"
    "io"
)

type (
    proto1 struct{}
)

func (p *proto1) SendRaw(w io.Writer, data []byte) (int, error) {
    return w.Write(data)
}

func (p *proto1) Read(r io.Reader) ([]byte, error) {
    header := make([]byte, 4)
    n, err := io.ReadFull(r, header)
    if err != nil {
        return nil, err
    }
    size := binary.BigEndian.Uint32(header)
    body := make([]byte, size)
    n, err = io.ReadFull(r, body)
    if err != nil {
        return nil, err
    }
    return body[:n], nil
}

func (p *proto1) Send(w io.Writer, data []byte) (int, error) {
    msg := p.Pack(data)
    return p.SendRaw(w, msg)
}

func (p *proto1) Pack(data []byte) []byte {
    header := make([]byte, 4)
    binary.BigEndian.PutUint32(header, uint32(len(data)))
    return append(header, data...)
}
