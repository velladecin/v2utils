package dns

import (
    "fmt"
    "context"
    "net"
    "time"
    "log"
    "regexp"
)

const (
    port  int = 53
    proto string = "udp"
)

type ResolverConfModifier func(rconf *ResolverConf)

type ResolverConf struct {
    Nameserver, Proto   string
    Port                int
}

type Resolver struct {
    Conf        *ResolverConf
    Resolver    *net.Resolver
}

func NewResolver(nameserver string, modifiers ...ResolverConfModifier) *Resolver {
    rconf := &ResolverConf{nameserver, proto, port}

    for _, mod := range modifiers {
        mod(rconf)
    }

    return &Resolver{
        Conf: rconf,
        Resolver: &net.Resolver{
            PreferGo: false,
            Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
                d := net.Dialer{
                    Timeout: time.Millisecond * time.Duration(10000),
                }

                return d.DialContext(ctx, rconf.Proto, fmt.Sprintf("%s:%d", rconf.Nameserver, rconf.Port))
            },
        },
    }
}

func (r *Resolver) Dig(hostname string) ([]string, error) {
    // This check is not great but I'm hoping that the user
    // has enough know-how to FQDN the hostname if/when needed
    if ok, _ := regexp.MatchString(`\.`, hostname); !ok {
        hostname = fmt.Sprintf("%s.v2.internal", hostname)
    }

    return r.Resolver.LookupHost(context.TODO(), hostname)
}

func (r *Resolver) Digx(ipaddr string) ([]string, error) {
    return r.Resolver.LookupAddr(context.TODO(), ipaddr)
}


//
// Updaters

func (r *Resolver) UpdateNameserver(nameserver string) {
    SetServer(nameserver)(r.Conf)
}

func (r *Resolver) UpdatePort(port int) {
    SetPort(port)(r.Conf)
}

func (r *Resolver) UpdateProto(proto string) {
    SetProto(proto)(r.Conf)
}


//
// Modifiers

func SetServer(server string) ResolverConfModifier {
    return func(rconf *ResolverConf) {
        rconf.Nameserver = server
    }
}

func SetProto(proto string) ResolverConfModifier {
    if proto != "tcp" && proto != "udp" {
        log.Fatalf("Invalid protocol(%s)", proto)
    }

    return func(rconf *ResolverConf) {
        rconf.Proto = proto
    }
}

func SetPort(port int) ResolverConfModifier {
    if port < 1 || port > 65535 {
        log.Fatalf("Port(%d) out of range", port)
    }

    return func(rconf *ResolverConf) {
        rconf.Port = port
    }
}
