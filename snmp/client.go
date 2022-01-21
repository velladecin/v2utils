package snmp

import (
    "github.com/soniah/gosnmp-master"
    "time"
    "regexp"
    "log"
    "os"
)

const (
    maxTimeout = 5

    // soniah/gosnmp.go
    // Set the number of retries to attempt
    maxRetries = 3

    // soniah/gosnmp.go
    // MaxRepetitions sets the GETBULK max-repetitions used by BulkWalk*
    // Unless MaxRepetitions is specified it will use defaultMaxRepetitions (50)
    // This may cause issues with some devices, if so set MaxRepetitions lower.
    // See comments in https://github.com/soniah/gosnmp/issues/100
    maxRepetitions uint8 = 40
)

type SnmpClient struct {
    conn                *gosnmp.GoSNMP
    target, community   string
    timeout, retries    int
    repetitions         uint8
    logger              gosnmp.Logger
}

type clientModifier func(*SnmpClient)

func NewSnmpClient(target string, mods ...clientModifier) (*SnmpClient, error) {
    ip4 := regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$`)
    ip6 := regexp.MustCompile(`(?i)^[a-f0-9\:]+$`)

    if ok4 := ip4.MatchString(target); !ok4 {
        if ok6 := ip6.MatchString(target); !ok6 {
            return nil, errInvalidTarget.Ctx("snmp client")
        }
    }

    c := &SnmpClient{
        target:         target,
        community:      "public",
        timeout:        maxTimeout,
        retries:        maxRetries,
        repetitions:    maxRepetitions, 
    }

    for _, m := range mods {
        m(c)
    }

    c.conn = &gosnmp.GoSNMP{
        Target:             c.target,
        Port:               161,
        Transport:          "udp",
        Community:          c.community,
        Version:            gosnmp.Version2c,
        Timeout:            time.Duration(c.timeout) * time.Second,
        Retries:            c.retries,
        ExponentialTimeout: false,
        // gosnmp.Default has this at 60, 40 should be enough for us..
        MaxOids:            40,
        // DO NOT SET THE BELOW TO: "c" => see warnings @ soniah/gosnmp.go
        // Caution: if you have set AppOpts to 'c', (Walk|BulkWalk)All() may loop indefinitely and cause an Out Of Memory
        AppOpts:            make(map[string]interface{}),
        MaxRepetitions:     c.repetitions,
        Logger:             c.logger,
    }

    err := c.conn.Connect()
    if err != nil {
        return nil, err
    }

    return c, nil
}

// Get()            - get single value
// Walk()           - get subtree of values using GETNEXT, request is made for each, returns slice of PDUs
// WalkStream()     - get subtree of values using GETNEXT, request is made for each, returns channel of (streaming) PDUs
// BulkWalk()       - get subtree of values using GETBULK, requests made in batches, returns slice of PDUs
// BulkWalkStream() - get subtree of values using GETBULK, requests made in batches, returns channel of (streaming) PDUs

type Pdu struct {
    Name, Type string
    Value interface{}
}

func marshallGosnmpPDU(pdu gosnmp.SnmpPDU) Pdu {
    p := Pdu{Name: pdu.Name}

    switch pdu.Type {
        case gosnmp.OctetString:                    p.Type = "byte"
        case gosnmp.ObjectIdentifier:               p.Type = "mib"
        case gosnmp.TimeTicks:                      p.Type = "timeticks"; p.Value = int(pdu.Value.(uint32))
        case gosnmp.Integer:                        p.Type = "int"
        case gosnmp.IPAddress:                      p.Type = "ipaddr"
        case gosnmp.Gauge32:                        p.Type = "gauge"    // single reading gives meaningful data
        case gosnmp.Counter32, gosnmp.Counter64:    p.Type = "counter"  // multiple (2+) readings to get meaningful data
        case gosnmp.Boolean:
        case gosnmp.BitString:
        case gosnmp.NoSuchObject:                   p.Type = "notfound"
        case gosnmp.EndOfContents:
        case gosnmp.EndOfMibView:
    }

    if p.Value == nil {
        p.Value = pdu.Value
    }

    return p
}

func (c *SnmpClient) Get(oids ...string) ([]Pdu, error) {
    p, err := c.conn.Get(oids)
    if err != nil {
        return []Pdu{}, err
    }

    // p = *gosnmp.SnmpPacket
    // &{Version:2c MsgFlags:0 SecurityModel:0 SecurityParameters:<nil> ContextEngineID: ContextName: Community:P0733Hhsca PDUType:162 MsgID:0 RequestID:479851905 MsgMaxSize:0 Error:NoError ErrorIndex:0 NonRepeaters:0 MaxRepetitions:0 Variables:[{Name:.1.3.6.1.2.1.1 Type:NoSuchObject Value:<nil>} {Name:.1.3.6.1.2.1.10.127.1.1.4.1.7.67788905 Type:OctetString Value:}] Logger:0xc000062280 SnmpTrap:{Variables:[] IsInform:false Enterprise: AgentAddress: GenericTrap:0 SpecificTrap:0 Timestamp:0}}

    var pdus []Pdu
    for _, pdu := range p.Variables {
        pdus = append(pdus, marshallGosnmpPDU(pdu))
    }

    return pdus, nil
}

func (c *SnmpClient) Walk(oid string) ([]Pdu, error) {
    p, err := c.conn.WalkAll(oid)
    if err != nil {
        return []Pdu{}, err
    }

    var pdus []Pdu
    for _, pdu := range p {
        pdus = append(pdus, marshallGosnmpPDU(pdu))
    }

    return pdus, nil
}

func (c *SnmpClient) BulkWalk(oid string) ([]Pdu, error) {
    p, err := c.conn.BulkWalkAll(oid)
    if err != nil {
        return []Pdu{}, err
    }

    var pdus []Pdu
    for _, pdu := range p {
        pdus = append(pdus, marshallGosnmpPDU(pdu))
    }

    return pdus, nil
}

func (c *SnmpClient) WalkStream(oids ...string) chan Pdu     { return c.walkAndStream(oids, false) }
func (c *SnmpClient) BulkWalkStream(oids ...string) chan Pdu { return c.walkAndStream(oids, true) }

func (c *SnmpClient) walkAndStream(oids []string, bulk bool) chan Pdu {
    stream := make(chan Pdu)
    fn := func(p gosnmp.SnmpPDU) error {
        stream <- marshallGosnmpPDU(p)
        return nil
    }

    go func() {
        for _, oid := range oids {
            var err error
            if bulk {
                err = c.conn.BulkWalk(oid, fn)
            } else {
                err = c.conn.Walk(oid, fn)
            }

            if err != nil {
                name := "WalkStream"
                if bulk { name = "BulkWalkStream" }

                stream <- Pdu{Name: name, Type: "error", Value: err.Error()}
                break
            }
        }

        close(stream)
    }()

    return stream
}


// SnmpClient modifiers

func Debug() clientModifier {
    return func(c *SnmpClient) {
        c.logger = log.New(os.Stdout, "", 0)
    }
}

func Community(community string) clientModifier {
    return func(c *SnmpClient) {
        c.community = community
    }
}

func Repetitions(repetitions uint8) clientModifier {
    if repetitions < 5 || repetitions > maxRepetitions {
        panic(errRepetitionsOutOfRange.Ctx("SNMP client config"))
    }

    return func(c *SnmpClient) {
        c.repetitions = repetitions
    }
}

func Timeout(timeout int) clientModifier {
    if timeout < 1 || timeout > maxTimeout {
        panic(errTimeoutOutOfRange.Ctx("SNMP client config"))
    }

    return func(c *SnmpClient) {
        c.timeout = timeout
    }
}

func Retries(retries int) clientModifier {
    if retries < 0 || retries > maxRetries {
        panic(errRetriesOutOfRange.Ctx("SNMP client config"))
    }

    return func(c *SnmpClient) {
        c.retries = retries
    }
}
