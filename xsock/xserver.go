package xsock

import (
    "net"
    "os"
    "io"
    "time"
    v2 "vella/v2utils"
)

const (
    sessTTL = 600 // 10mins
)

type session struct {
    time    time.Time
    count   int
}

type sessions map[string]*session

type serverModifier func(*Server)

type Server struct {
    Path        string
    Listener    net.Listener
    Force       bool
    Sessions    sessions
    Ttl         int
    Handler     func(string)(string, bool) // func(question)(answer, exit)
}

func NewServer(path string, mods ...serverModifier) (*Server, error) {
    s := &Server{Path: path, Sessions: make(sessions), Ttl: sessTTL}

    for _, m := range mods {
        m(s)
    }

    if fi, err := os.Stat(path); err == nil {
        if s.Force && (fi.Mode() & os.ModeSocket) != 0 {
            os.Remove(path) // stale socket
        }
    }

    l, err := net.Listen("unix", path)
    if err != nil {
        return nil, v2.ErrCtx(err.Error(), "xserver")
    }

    s.Listener = l
    return s, nil
}

func (s *Server) AcceptAndHandle() error {
    if s.Handler == nil {
        panic(errMissingHandler.Ctx("xserver"))
    }

    conn, err := s.Listener.Accept()
    if err != nil {
        return v2.ErrCtx(err.Error(), "xserver")
    }
    defer conn.Close()

    p, err := rx(conn)
    if err != nil {
        if err == io.EOF { // close()
            return nil
        }

        return v2.ErrCtx(err.Error(), "xserver")
    }

    switch ; {

        // SYN
        // generate session ID, update packet+session with ID, set to ACK and return

        case p.IsSyn():
            s.RegisterSession(p.SetId())
            p.SetAck()

        // CLS
        case p.IsCls():
            s.UnregisterSession(p.GetId())
            return nil

        // ACK
        // any other receives are ACK and expect answer

        default:
            var response string
            var exitcode bool

            err = s.Sessions.update(p.GetIdStr(), s.Ttl)
            if err != nil {
                response, exitcode = err.Error(), false
            } else {
                response, exitcode = s.Handler(p.GetMsg())
            }

            p.ResetMsg(response)
            p.SetExitCode(exitcode)
    }

    return tx(conn, p)
}

func (s *Server) UnregisterSession(id [idLen]byte) sessions {
    sid := string(id[:])

    r := make(sessions)
    r[sid] = s.Sessions[sid]

    delete(s.Sessions, sid)

    return r
}

func (s *Server) RegisterSession(id []byte) {
    s.Sessions[string(id)] = &session{time: time.Now(), count: 0}
}

func (s *Server) RegisterHandler(fn func(string)(string, bool)) {
    s.Handler = fn
}

func (s sessions) update(sid string, ttl int) error {
    if _, ok := s[sid]; !ok {
        return errInvalidId.Ctx("xserver")
    }

    var err error
    var del []string

    now := time.Now()
    for id, val := range s {
        if now.Sub(val.time) > time.Duration(ttl) * time.Second {
            if sid == id {
                err = errExpiredId.Ctx("xserver")
            }

            del = append(del, id)
        }
    }

    if err == nil && len(del) > 0 {
        panic(v2.ErrCtx("How did I get here??", "xserver"))
    }

    if err != nil {
        for _, id := range del {
            delete(s, id)
        }

        return err
    }

    s[sid].count++
    return nil
}


//
// Modifiers

func Force(b bool) serverModifier {
    return func(s *Server) {
        s.Force = b
    }
}

func SessTTL(ttl int) serverModifier {
    if ttl > 3600 || ttl < 1 {
        panic(errTTL.Ctx("xserver"))
    }

    return func(s *Server) {
        s.Ttl = ttl
    }
}
