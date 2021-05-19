package common

import (
    "io"
)

type protocol interface {
    Read(r io.Reader) ([]byte, error)
    Send(w io.Writer, data []byte) (int, error)
    SendRaw(w io.Writer, data []byte) (int, error)
    Pack(data []byte) []byte
}

func newProtocol(pType int) protocol {
    switch pType {
    case 1:
        return &proto1{}
    default:
        return &proto0{}
    }
}
