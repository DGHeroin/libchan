package libchan

import (
    "context"
    "io"
)

type (
    Closer interface {
        Close() error
    }
    Chan interface {
        io.ReadWriteCloser
        ReadMessage() ([]byte, error)
    }
    Transport interface {
        Accept() Chan
        Dial() (Chan, error)
        Context() context.Context
        Close() error
    }
    Attr map[string]string
)
