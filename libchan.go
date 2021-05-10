package libchan

import "context"

type (
    Transport interface {
        Send([]byte) error
        Recv() ([]byte, error)
        Context() context.Context
    }
    Attr map[string]string
)
