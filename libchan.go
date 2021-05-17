package libchan

import "context"

type (
    Closer interface {
        Close() error
    }
    Chan interface {
        Send([]byte) error
        Read() ([]byte, error)
        Close() error
    }
    Transport interface {
        Accept() Chan
        Dial() (Chan, error)
        Context() context.Context
        Close() error
    }
    Attr map[string]string
)
