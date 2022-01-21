package http

import (
    "fmt"
    "log"
    "net/http"
    "io/ioutil"
    "encoding/json"
    "strings"
)

type HttpClientModifier func(hc *HttpClient)

type HttpClient struct {
    Schema, Server string
    Port int
}

func NewHttpClient(server string, mods ...HttpClientModifier) *HttpClient {
    hc := &HttpClient{
        Schema: "http",
        // Won't attempt to check hostname validity
        // as that's none of my business..
        Server: server,
        Port: 80,
    }

    for _, m := range mods {
        m(hc)
    }

    return hc
}

func (hc *HttpClient) GetUri(uri string) ([]byte, error) {
    url := fmt.Sprintf("%s://%s:%d/%s", hc.Schema, hc.Server, hc.Port, uri)

    resp, err := http.Get(url)
    if err != nil {
        return []byte{}, err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return []byte{}, err
    }

    return body, nil
}

func (hc *HttpClient) GetUriJson(uri string, i interface{}) error {
    body, err := hc.GetUri(uri)
    if err != nil {
        return err
    }

    if err := json.Unmarshal(body, i); err != nil {
        return err
    }

    return nil
}


//
// Modifiers

func SetSchema(schema string) HttpClientModifier {
    schema = strings.ToLower(schema)

    if schema != "http" && schema != "https" {
        log.Fatalf("Schema must be HTTP(s), not(%s)", schema) 
    }

    return func(hc *HttpClient) {
        hc.Schema = schema
    }
}

func SetPort(port int) HttpClientModifier {
    if port < 1 || port > 65535 {
        log.Fatalf("Port must be within range 1-65535, not(%d)", port)
    }

    return func(hc *HttpClient) {
        hc.Port = port
    }
}
