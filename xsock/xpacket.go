package xsock

import (
    "math/rand"
    v2 "vella/v2utils"
)

const (
    // full packet
    packetLen       = 512

    // id
    idLen           = 32
    idPool          = "AaBbCcDdEeFfGgHhIiJjKkLlMmNnOoPpQqRrSsTtUuVvWwXxYyZz0123456789"

    // session
    sessLen         = 2
    syn uint8       = 0 // type syn
    ack uint8       = 1 // type ack
    cls uint8       = 2 // type close
    scs uint8       = 0 // exit success
    flt uint8       = 1 // exit fail

    // message
    msgLen          = packetLen - idLen - sessLen
)

type pktMod func(*Packet)

type Packet struct {
    Stream      [packetLen]byte
}

func NewPacket(pm ...pktMod) *Packet {
    p := &Packet{}

    // SYN is default
    // ACK, CLS possible modifies (cls does not make much sense though..)

    if len(pm) > 0 {
        pm[0](p)
    }

    return p
}


//
// Isers

func (p *Packet) IsSyn() bool { return p.Stream[idLen] == syn }
func (p *Packet) IsAck() bool { return p.Stream[idLen] == ack }
func (p *Packet) IsCls() bool { return p.Stream[idLen] == cls }

func (p *Packet) IsCorrupt() (bool, error) {
    // total length - we care
    // id           - we care
    // sess.type    - we care
    // sess.exit    - we don't care
    // msg          - we don't care

    if len(p.Stream) != packetLen {
        return true, errLength.Ctx("xpacket")
    }

    if !p.IsSyn() && !p.IsAck() && !p.IsCls() {
        return true, errType.Ctx("xpacket")
    }

    if err := ValidHeaderId(p.GetId()); err != nil {
        return true, v2.ErrCtx(err.Error(), "xpacket")
    }

    return false, nil
}


//
// Getters

func (p *Packet) GetIdStr() string {
    id := p.GetId()
    return string(id[:])
}

func (p *Packet) GetId() [idLen]byte {
    var i [idLen]byte

    for c, b := range p.Stream {
        if c == idLen {
            break
        }

        i[c] = b
    }

    return i
}

func (p *Packet) GetSess() [sessLen]byte {
    var i [sessLen]byte

    for c, sc:=idLen, 0; c<idLen+sessLen; c++ {
        i[sc] = p.Stream[c]
        sc++
    }

    return i
}

func (p *Packet) GetMsg() string {
    // here we only expect ASCII chars,
    // is this good enough..?

    b := []byte{}
    for c:=idLen+sessLen; c<packetLen; c++ {
        if p.Stream[c] == 0 {
            break
        }

        b = append(b, p.Stream[c])
    }

    return string(b)    
}


//
// Setters

func (p *Packet) SetExitCode(b bool) {
    // true = success
    // false = fail

    var e byte = flt
    if b {
        e = scs
    }

    p.Stream[idLen+sessLen-1] = e
}

func (p *Packet) ResetMsg(s string) {
    for c:=idLen+sessLen; c<packetLen; c++ {
        p.Stream[c] = 0
    }

    p.SetMsg(s)
}

func (p *Packet) SetMsg(s string) {
    for c, b := range s {
        p.Stream[idLen+sessLen+c] = byte(b)
    }
}

func (p *Packet) SetSyn() { p.Stream[idLen] = syn }
func (p *Packet) SetAck() { p.Stream[idLen] = ack }
func (p *Packet) SetCls() { p.Stream[idLen] = cls }

func (p *Packet) SetId(id ...pktMod) []byte {
    var pim pktMod // pkt id modifier

    switch ; {
        case len(id) > 0:
            pim = id[0] // allow max one ID
        default:
            // Id string generation => https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go/22892986#22892986
            // Id 3 - Remainder
            var b [idLen]byte
            for c:=0; c<idLen; c++ {
                b[c] = idPool[rand.Int63() % int64(len(idPool))]
            }

            pim = setId(b)
    }

    pim(p) // pimpin'..
    return p.Stream[:idLen]
}


//
// Generic

func ValidHeaderId(id [idLen]byte) error {
    // two options
    // 1. SYN must be all byte(0)
    // 2. ACK must be from idpoolmap

    switch id[0] {
        case 0:
            for _, b := range id {
                if b != 0 {
                    return errHeader.Ctx("xpacket(syn)")
                }
            }
        default:
            // lookup map
            idpoolmap := make(map[byte]byte)
            for _, b := range idPool {
                idpoolmap[byte(b)] = uint8(1)
            }

            for _, b := range id {
                if _, ok := idpoolmap[b]; !ok {
                    return errHeader.Ctx("xpacket(ack)")
                }
            }
    }

    return nil
}


//
// Packet modifiers

func setId(b [idLen]byte) pktMod {
    return func(p *Packet) {
        for c:=0; c<len(b); c++ {
            p.Stream[c] = b[c]
        }
    }
}

func Ack(sessid [idLen]byte) pktMod {
    return func(p *Packet) {
        p.SetId(setId(sessid)) 
        p.SetAck()
    }
}

func Cls(sessid [idLen]byte) pktMod {
    return func(p *Packet) {
        p.SetId(setId(sessid))
        p.SetCls()
    }
}
