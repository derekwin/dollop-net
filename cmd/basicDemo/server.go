package main

import (
	"context"

	"github.com/derekwin/dollop-net/dollop"
	dtls "github.com/derekwin/dollop-net/dollop/tls"
)

const testaddr = "127.0.0.1:19999"

func main() {
	tlsServer, err := dtls.CreateServerTLSConfig(testaddr, "../certs/server.crt", "../certs/server.key", true)
	if err != nil {
		panic(err)
	}

	server, err := dollop.NewServer("test", dollop.WithTlsConfig(tlsServer))
	if err != nil {
		panic(err)
	}

	server.Serve(context.Background(), testaddr)
}
