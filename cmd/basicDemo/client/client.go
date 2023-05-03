package main

import (
	"context"
	"fmt"
	"io"

	"github.com/derekwin/dollop-net/dollop"
	"github.com/derekwin/dollop-net/dollop/frame"
	dtls "github.com/derekwin/dollop-net/dollop/tls"
	"github.com/quic-go/quic-go"
)

const testaddr = "127.0.0.1:19999"

func handler(ctx context.Context, stream quic.Stream, data []byte) {
	fmt.Printf("Client: Sending '%s' to stream %d\n", data, stream.StreamID())
	cnt, err := stream.Write(data)
	if err != nil {
		panic(err)
	}
	fmt.Println("write success ", cnt)

	buf := make([]byte, len(data))
	_, err = io.ReadFull(stream, buf)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Client: Got '%s'\n", buf)
}

func main() {
	tlsClient, err := dtls.CreateClientTLSConfig("../../certs/client.crt", "../../certs/client.key", true)
	if err != nil {
		panic(err)
	}

	conn, err := quic.DialAddr(testaddr, tlsClient, dollop.DefalutQuicConfig)
	if err != nil {
		panic(err)
	}

	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(11)
	controlStream := dollop.NewFrameStream(stream)

	controlStream.WriteFrame(frame.NewRequestDataStreamFrame(frame.StreamID(stream.StreamID())))

	// for {
	// 	dataStream, err := conn.AcceptStream(conn.Context())
	// 	if err != nil {
	// 		return
	// 	}
	// }
	fmt.Println("request new data stream, awaiting")
	datastream, err := conn.AcceptStream(conn.Context())
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("new data stream", datastream.StreamID())
	data := []byte("data from clienttttt")
	handler(context.Background(), datastream, data)
}
