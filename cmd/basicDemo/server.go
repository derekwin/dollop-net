package main

import (
	"context"
	"fmt"

	"github.com/derekwin/dollop-net/dollop"
	"github.com/derekwin/dollop-net/dollop/frame"
	dtls "github.com/derekwin/dollop-net/dollop/tls"
)

const testaddr = "127.0.0.1:19999"

type RawStreamRouter struct {
	dollop.BaseRawRouter
}

func (lr RawStreamRouter) Handler(req dollop.RawRequestI) {
	stream := req.GetStream()
	fmt.Printf("--- handler raw data '%s' from id %d \n", req.GetData(), stream.StreamID())
	_, err := stream.Write(req.GetData())
	if err != nil {
		fmt.Println(err)
	}
}

type FrameStreamRouter struct {
	dollop.BaseFrameRouter
}

func (lr FrameStreamRouter) Handler(req dollop.FrameRequestI) {
	stream := req.GetStream()
	fmt.Printf("--- handler frame data '%s' from id %d \n", string(req.GetData()), stream.StreamID())
	err := stream.WriteFrame(frame.NewDataFrame(req.GetData()))
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	tlsServer, err := dtls.CreateServerTLSConfig(testaddr, "../certs/server.crt", "../certs/server.key", true)
	if err != nil {
		panic(err)
	}

	rawrouter := RawStreamRouter{}
	framerouter := FrameStreamRouter{}

	server, err := dollop.NewServer("test", dollop.WithTlsConfig(tlsServer),
		dollop.WithRawRouter(rawrouter), dollop.WithFrameRouter(framerouter))
	if err != nil {
		panic(err)
	}

	server.Serve(context.Background(), testaddr)
}
