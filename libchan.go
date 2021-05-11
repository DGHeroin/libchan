package libchan

import "context"

type (
    Chan interface {
        Send([]byte) error
        SendBatching([]byte) error
        Read() ([]byte, error)
    }
    Transport interface {
        Accept() Chan
        Dial() (Chan, error)
        Context() context.Context
    }
    Attr map[string]string
)
