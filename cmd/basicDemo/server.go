package main

import (
	"context"
	"fmt"

	"github.com/derekwin/dollop-net/dollop"
	dtls "github.com/derekwin/dollop-net/dollop/tls"
)

const testaddr = "127.0.0.1:19999"

type RawStreamRouter struct {
	dollop.BaseRawRouter
}

func (lr RawStreamRouter) Handler(req dollop.RawRequestI) error {
	stream, err := req.GetStream()
	if err != nil {
		fmt.Println(err)
		return err
	}
	data, err := req.GetData()
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Printf("--- handler raw data '%s' from id %d \n", data, stream.StreamID())
	_, err = stream.Write(data)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func main() {
	tlsServer, err := dtls.CreateServerTLSConfig(testaddr, "../certs/server.crt", "../certs/server.key", true)
	if err != nil {
		panic(err)
	}

	rawrouter := RawStreamRouter{}

	server, err := dollop.NewServer("test", dollop.WithTlsConfig(tlsServer),
		dollop.WithRawRouter(rawrouter))
	if err != nil {
		panic(err)
	}

	server.Serve(context.Background(), testaddr)
}
