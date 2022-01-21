package xsock

import (
    "net"
    "time"
)

type Client struct {
    Path    string
    Session [idLen]byte
}

func NewClient(path string) (*Client, error) {
    c := &Client{Path: path}

    p, e := c.transmit()
    if e != nil {
        return nil, e
    }

    c.SetSession(p)
    return c, nil
}

func (c *Client) Send(s string) (string, error) {
    p, err := c.transmit(s)
    if err != nil {
        return "", err
    }

    if !c.ActiveSession() {
        c.SetSession(p)
    }

    return p.GetMsg(), nil
}

type cSession struct {
    p *Packet
    e error
}

func (c *Client) transmit(s ...string) (*Packet, error) {
    conn, err := net.Dial("unix", c.Path)
    if err != nil {
        return nil, err
    }
    defer conn.Close()

    // NewPacket() returns SYN
    // if active session make it ACK

    p := NewPacket()
    if c.ActiveSession() {
        Ack(c.Session)(p)
        p.SetMsg(s[0])
    }

    rcv := make(chan cSession)
    go func() {
        p, e := rx(conn)
        rcv <- cSession{p, e}           // reply listener
        close(rcv)
    }()

    err = tx(conn, p)                   // sender
    if err != nil {
        return nil, err
    }

    var sess cSession
    select {
        case sess = <- rcv:             // response received (from reply listener, blocking channel - timeout below)
        case <- time.After(time.Duration(3) * time.Second):
            return nil, errTransmitTimeout.Ctx("xclient")
    }

    return sess.p, sess.e
}

func (c *Client) SetSession(p *Packet) {
    for count:=0; count<int(idLen); count++ {
        c.Session[count] = p.Stream[count]
    }    
}

func (c *Client) ActiveSession() bool {
    // is ValidHeaderId(c.Session) needed here,
    // we should either have none or valid one..?

    if c.Session[0] == 0 {
        return false
    }

    return true
}

func (c *Client) Close() error {
    conn, err := net.Dial("unix", c.Path)
    if err != nil {
        return err
    }
    defer conn.Close()

    p := NewPacket()
    Cls(c.Session)(p)

    err = tx(conn, p)
    if err != nil {
        return err
    }

    c.Session = [idLen]byte{}
    return nil
}
