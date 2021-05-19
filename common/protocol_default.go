package common

import (
    "encoding/binary"
    "fmt"
    "io"
)

type (
    proto0 struct {}
)

func (p *proto0) SendRaw(w io.Writer, data []byte) (int, error) {
    return w.Write(data)
}

func (p *proto0) Read(r io.Reader) ([]byte, error) {
    header := make([]byte, 5)
    n, err := io.ReadFull(r, header)
    if err != nil {
        return nil, err
    }
    msgT := header[4]
    switch msgT {
    case 0x98:
        size := binary.BigEndian.Uint32(header)
        body := make([]byte, size)
        n, err = io.ReadFull(r, body)
        if err != nil {
            return nil, err
        }
        return body[:n], nil
    default:
        return nil, fmt.Errorf("unsupport msg type")
    }
}

func (p *proto0) Send(w io.Writer, data []byte) (int, error) {
    msg := p.Pack(data)
    return p.SendRaw(w, msg)
}

func (p *proto0) Pack(data []byte) []byte {
    header := make([]byte, 5)
    header[4] = 0x98
    binary.BigEndian.PutUint32(header, uint32(len(data)))
    return append(header, data...)
}
