package dollop

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/derekwin/dollop-net/dollop/frame"
	"github.com/quic-go/quic-go"
)

type StreamID int64

const ServerConnectionCloseCode quic.ApplicationErrorCode = 701

type ConnectionIS interface {
	GetConn()
	setControlStream()
	addDataStream()
	deleteDataStream()
	Close()

	// for Server connection
	Serve()
	controlStreamLoop()
}

type ConnectionIC interface {
	GetConn()
	setControlStream()
	addDataStream()
	deleteDataStream()
	Close()

	// for Client connection
	Connect()
}

type ServerConnection struct {
	ctx                   context.Context
	conn                  quic.Connection
	controlStream         FrameStream // control stream
	requestDataStreamChan chan *frame.RequestDataStreamFrame
	streams               sync.Map // data stream
	group                 sync.WaitGroup
}

func NewServerConnection(ctx context.Context, conn quic.Connection) *ServerConnection {
	return &ServerConnection{ctx: ctx, conn: conn, requestDataStreamChan: make(chan *frame.RequestDataStreamFrame, 10)}
}

func handler(stream quic.Stream) error {
	fmt.Println("data from stream id ", stream.StreamID())

	buf := make([]byte, 512)
	_, err := stream.Read(buf[:])
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("handler data %s \n", buf)
	_, err = stream.Write(buf[:])
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (cs *ServerConnection) Serve(ctx context.Context) <-chan struct{} {
	done := make(chan struct{})

	qconn := cs.conn

	// wg := new(sync.WaitGroup)

	controlStream, err := qconn.AcceptStream(ctx)
	if err != nil {
		fmt.Println(err)
	}
	defer controlStream.Close()

	cs.setControlStream(NewFrameStream(controlStream))
	go cs.controlStreamLoop()

	go func(cs *ServerConnection) {

		defer cs.Close()

		select {
		case <-ctx.Done():
			return
		case <-cs.streamManager(ctx):
		}
	}(cs)

	return done
}

func (cs *ServerConnection) controlStreamLoop() {
	for {
		f, err := cs.controlStream.ReadFrame()
		if err != nil {
			cs.conn.CloseWithError(0, err.Error())
			return
		}
		switch ff := f.(type) {
		case frame.RequestDataStreamFrame:
			cs.requestDataStreamChan <- &ff
		default:
			fmt.Println("control stream read unexcepted frame", "frame_type", f.Type())
		}
	}
}

func (cs *ServerConnection) streamManager(ctx context.Context) chan struct{} {
	for {
		ff, ok := <-cs.requestDataStreamChan
		if !ok {
			fmt.Println(io.EOF)
		}

		// err := handshake(ff)

		// 如果出错，返回拒绝
		// cs.controlStream.WriteFrame(frame.NewRejectDataStreamFrame(frame.StreamID(cs.controlStream.stream.StreamID())))
		fmt.Println("receive request", ff)
		newStream, err := cs.conn.OpenStreamSync(ctx)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("open new stream", newStream.StreamID())
		_, err = newStream.Write(frame.NewAckDataStreamFrame(frame.StreamID(cs.controlStream.stream.StreamID())).Encode())
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("write ack to new stream")

		cs.streams.Store(ff.StreamId(), newStream)

		// 将数据请求封装为request，然后分别调用对应的router
		go cs.processStream(newStream)
	}
}

func (cs *ServerConnection) processStream(stream quic.Stream) {
	fmt.Println("process new stream", stream.StreamID())
	handler(stream)
}

func (cs *ServerConnection) Close() {

	cs.controlStream.Close()

	cs.streams.Range(func(key, value interface{}) bool {
		stream := value.(quic.Stream)
		stream.Close()
		return true
	})

	cs.conn.CloseWithError(ServerConnectionCloseCode, "server side close this conn")
}

func (cs *ServerConnection) GetConn() quic.Connection {
	return cs.conn
}

func (cs *ServerConnection) setControlStream(stream FrameStream) {
	cs.controlStream = stream
}

func (cs *ServerConnection) addDataStream(id StreamID, stream quic.Stream) {
	cs.streams.Store(id, stream)
}

func (cs *ServerConnection) deleteStream(id StreamID) {
	cs.streams.Delete(id)
}

func (cs *ServerConnection) Wait() { cs.group.Wait() }
