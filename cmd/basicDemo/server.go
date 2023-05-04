package main

import (
	"context"
	"fmt"

	"github.com/derekwin/dollop-net/dollop"
	dtls "github.com/derekwin/dollop-net/dollop/tls"
)

const testaddr = "127.0.0.1:19999"

type LocalRouter struct {
	dollop.BaseRouter
}

func (lr LocalRouter) Handler(req dollop.RequestI) {
	stream := req.GetStream()
	fmt.Printf("--- handler data '%s' from id %d \n", req.GetData(), stream.StreamID())
	_, err := stream.Write(req.GetData())
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	tlsServer, err := dtls.CreateServerTLSConfig(testaddr, "../certs/server.crt", "../certs/server.key", true)
	if err != nil {
		panic(err)
	}

	router := LocalRouter{}

	server, err := dollop.NewServer("test", dollop.WithTlsConfig(tlsServer),
		dollop.WithRouter(router))
	if err != nil {
		panic(err)
	}

	server.Serve(context.Background(), testaddr)
}
