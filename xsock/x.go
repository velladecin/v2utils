package xsock

import (
    "net"
    "io"
    v2 "vella/v2utils"
)

func rx(c net.Conn) (*Packet, error) {
    buf := make([]byte, packetLen)

    _, err := c.Read(buf)
    if err != nil { 
        if err != io.EOF {
            err = v2.ErrCtx(err.Error(), "rx")
        }

        return nil, err
    }           

    if len(buf) > packetLen {
        return nil, errBufferOverflow.Ctx("rx")
    }

    p := NewPacket()
    for c, v := range buf {
        p.Stream[c] = v
    }

    if _, err = p.IsCorrupt(); err != nil {
        return nil, errCorrupted.Ctx("rx")
    }

    return p, nil
}

func tx(c net.Conn, p *Packet) error {
    _, err := c.Write(p.Stream[:])
    if err != nil {
        return v2.ErrCtx(err.Error(), "tx")
    }

    return nil
}
