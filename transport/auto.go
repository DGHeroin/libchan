package transport

import (
    "github.com/DGHeroin/libchan"
    "github.com/DGHeroin/libchan/kcp"
    "github.com/DGHeroin/libchan/tcp"
    "net/url"
)

func New(rawurl string) libchan.Transport {
    u, err := url.Parse(rawurl)
    if err != nil {
        return nil
    }
    switch u.Scheme {
    case "kcp":
        return kcp.New(rawurl)
    case "tcp":
        return tcp.New(rawurl)
    default:
        return nil
    }
}
