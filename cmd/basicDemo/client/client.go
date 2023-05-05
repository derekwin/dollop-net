package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/derekwin/dollop-net/dollop"
	"github.com/derekwin/dollop-net/dollop/frame"
	dtls "github.com/derekwin/dollop-net/dollop/tls"
)

func handler(ctx context.Context, stream dollop.RawStreamI, data []byte) {
	fmt.Printf("Client: Sending '%s' to stream %d\n", data, stream.StreamID())
	cnt, err := stream.Write(data)
	if err != nil {
		panic(err)
	}
	fmt.Println("write success ", cnt)

	buf := make([]byte, 512)
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

	client := dollop.NewClient("testclient", tlsClient, dollop.DefalutQuicConfig)

	client.Connect("127.0.0.1:19999")

	datastream, _, err := client.NewRawStream()
	if err != nil {
		log.Fatal(err)
	}

	data := []byte("data from clienttttt")
	handler(context.Background(), datastream, data)

	time.Sleep(time.Second * 2)
	datastream2, _, err := client.NewRawStream()
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Second * 2)
	handler(context.Background(), datastream, data)
	time.Sleep(time.Second * 2)
	handler(context.Background(), datastream2, data)

	framestream, _, err := client.NewFrameStream()
	if err != nil {
		log.Fatal(err)
	}
	framestream.WriteFrame(frame.NewDataFrame(data))
	f, err := framestream.ReadFrame()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(f.GetData()))
}
