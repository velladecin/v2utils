package v2utils

import (
    "fmt"
    "regexp"
    "strings"
)

const (
    portStart = 1
    portHigh = 1024
    portEnd = 65535
)

var (
    errInvalidPort          = Err("invalid port")
    errInvalidMacAddr       = Err("invalid mac address")
    errInvalidMacAddrFormat = Err("invalid mac address format")
)

type MacAddr string

func NewMacAddr(macaddr string) (MacAddr, error) {
    mpool := regexp.MustCompile(`(?i)[\.\:a-f0-9]`)

    var mac []byte
    for _, c := range macaddr {
        c_str := string(c)

        if ok := mpool.MatchString(c_str); !ok {
            return "", errInvalidMacAddr
        }

        if c_str == "." || c_str == ":" {
            continue
        }

        mac = append(mac, byte(c))
    }

    m := MacAddr(mac)

    err := m.validate()
    if err != nil {
        return MacAddr(""), err
    }

    return m, nil
}

func (m MacAddr) validate() error {
    maddr := regexp.MustCompile(`(?i)^[a-f0-9]{12}$`)

    if ok := maddr.MatchString(string(m)); !ok {
        return errInvalidMacAddr
    }

    return nil
}

func (m MacAddr) ToUpper() MacAddr { return m.fontcase("upper") }
func (m MacAddr) ToLower() MacAddr { return m.fontcase("lower") }
func (m MacAddr) fontcase(c string) MacAddr {
    if err := m.validate(); err != nil {
        panic(err)
    }

    var mm MacAddr
    switch c {
        case "upper": mm = MacAddr(strings.ToUpper(string(m)))
        case "lower": mm = MacAddr(strings.ToLower(string(m)))
    }

    return mm
}

func (m MacAddr) Getf(format string) (string, error) {
    if err := m.validate(); err != nil {
        panic(err)
    }

    var s string
    switch format {
        // E6000 Arris CMTS format aaaa.0099.ffff
        case "e6000":
            s = fmt.Sprintf("%s.%s.%s", m[0:4], m[4:8], m[8:12])
        case "lean", "short":
            s = string(m)
        case "long":
            s = fmt.Sprintf("%s:%s:%s:%s:%s:%s", m[0:2], m[2:4], m[4:6], m[6:8], m[8:10], m[10:12])
        default:
            return s, errInvalidMacAddrFormat
    }

    return s, nil
}

func (m MacAddr) GetFullRgxString() string {
    // Getf() will do validate()

    m1, _ := m.Getf("e6000")
    m2, _ := m.Getf("lean")
    m3, _ := m.Getf("long")

    return fmt.Sprintf("%s|%s|%s", m1, m2, m3)
}

func LowPort(port int) (bool, error) {
    err := PortRange(port)
    if err != nil {
        return false, err
    }

    var ok bool = false
    if port < portHigh {
        ok = true
    }

    return ok, nil
}

func HighPort(port int) (bool, error) {
    err := PortRange(port)
    if err != nil {
        return false, err
    }

    var ok bool = false
    if port >= portHigh {
        ok = true
    }

    return ok, nil
}

func PortRange(port int) error {
    if port < portStart || port > portEnd {
        return errInvalidPort
    }

    return nil
}
