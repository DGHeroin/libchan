package common

import (
    "bytes"
    "compress/gzip"
    "io/ioutil"
)

func doCompression(t int, data []byte) []byte {
    switch t {
    case 1:
        r, _ := gzipCompress(data)
        return r
    default:
        return data
    }
}

func doUnCompression(t int, data []byte) []byte {
    switch t {
    case 1:
        r, _ := gzipdecompress(data)
        return r
    default:
        return data
    }
}

func gzipCompress(data []byte) ([]byte, error) {
    var b bytes.Buffer
    gz := gzip.NewWriter(&b)
    if _, err := gz.Write([]byte(data)); err != nil {
        return nil, err
    }
    if err := gz.Close(); err != nil {
        return nil, err
    }
    return b.Bytes(), nil
}

func gzipdecompress(input []byte) ([]byte, error) {
    gr, err := gzip.NewReader(bytes.NewBuffer(input))
    defer gr.Close()
    data, err := ioutil.ReadAll(gr)
    if err != nil {
        return nil, err
    }
    return data, nil
}
