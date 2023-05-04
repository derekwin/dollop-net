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

type ConnectionI interface {
	GetConn() quic.Connection
	GetStream(id StreamID) (quic.Stream, error)
	setControlStream(stream *FrameStream)
	addDataStream(id StreamID, stream quic.Stream)
	deleteStream(id StreamID)
	Close()
}

type Connection struct {
	ctx                   context.Context
	conn                  quic.Connection
	controlStream         *FrameStream // control stream
	requestDataStreamChan chan *frame.RequestDataStreamFrame
	streams               sync.Map // data stream
	group                 sync.WaitGroup
}

func newConnection(ctx context.Context) Connection {
	return Connection{ctx: ctx, requestDataStreamChan: make(chan *frame.RequestDataStreamFrame, 10)}
}

func (c *Connection) Close() {

	c.controlStream.Close()

	c.streams.Range(func(key, value interface{}) bool {
		stream := value.(quic.Stream)
		stream.Close()
		return true
	})

	c.conn.CloseWithError(ServerConnectionCloseCode, "server side close this conn")
}

func (c *Connection) GetConn() quic.Connection {
	return c.conn
}

func (c *Connection) GetStream(id StreamID) (quic.Stream, error) {
	stream, ok := c.streams.Load(id)
	if !ok {
		return nil, fmt.Errorf("err get valid stream")
	}
	return stream.(quic.Stream), nil
}

func (c *Connection) setControlStream(stream *FrameStream) {
	c.controlStream = stream
}

func (c *Connection) addDataStream(id StreamID, stream quic.Stream) {
	c.streams.Store(id, stream)
}

func (c *Connection) deleteStream(id StreamID) {
	c.streams.Delete(id)
}

func (c *Connection) Wait() { c.group.Wait() }

// Server Connection impliment specisal
type ConnectionIS interface {
	ConnectionI
	// for Server connection
	Serve()
	bindRouters([]RouterI)
	controlStreamLoop()
}

type ServerConnection struct {
	*Connection
	Routers []RouterI
}

func NewServerConnection(ctx context.Context, qconn quic.Connection) *ServerConnection {
	conn := newConnection(ctx)

	// init other value
	conn.conn = qconn

	return &ServerConnection{Connection: &conn}
}

func (sc *ServerConnection) Serve(ctx context.Context) <-chan struct{} {
	done := make(chan struct{})

	qconn := sc.conn

	// wg := new(sync.WaitGroup)

	controlStream, err := qconn.AcceptStream(ctx)
	if err != nil {
		fmt.Println(err)
	}
	defer controlStream.Close()

	sc.setControlStream(NewFrameStream(controlStream))
	go sc.controlStreamLoop()

	go func(sc *ServerConnection) {

		defer sc.Close()

		select {
		case <-ctx.Done():
			return
		case <-sc.streamManager(ctx):
		}
	}(sc)

	return done
}

func (sc *ServerConnection) controlStreamLoop() {
	for {
		f, err := sc.controlStream.ReadFrame()
		if err != nil {
			sc.conn.CloseWithError(0, err.Error())
			return
		}
		switch ff := f.(type) {
		case frame.RequestDataStreamFrame:
			sc.requestDataStreamChan <- &ff
		default:
			fmt.Println("control stream read unexcepted frame", "frame_type", f.Type())
		}
	}
}

func (sc *ServerConnection) streamManager(ctx context.Context) chan struct{} {
	for {
		ff, ok := <-sc.requestDataStreamChan
		if !ok {
			fmt.Println(io.EOF)
		}

		// err := handshake(ff)

		// 如果出错，返回拒绝
		// cs.controlStream.WriteFrame(frame.NewRejectDataStreamFrame())

		fmt.Println("receive new stream request", ff)
		newStream, err := sc.conn.OpenStreamSync(ctx)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("open new stream", newStream.StreamID())
		_, err = newStream.Write(frame.NewAckDataStreamFrame().Encode())
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("ack to new stream")

		sc.addDataStream(StreamID(newStream.StreamID()), newStream)

		go sc.processStream(newStream)
	}
}

func (sc *ServerConnection) processStream(stream quic.Stream) {
	fmt.Println("process stream", stream.StreamID())
	buf := make([]byte, 512) // 分配一次，重复使用
	for {
		// 判断ctx业务退出? 是否有必要

		// 读取数据
		_, err := stream.Read(buf[:])
		if err != nil {
			// fmt.Println(err) // 客户端退出后，会触发超时
			break
		}
		// 将数据请求封装为request，然后分别调用对应的router
		// 生成request
		req := &Request{conn: sc, stream: stream, data: buf} // 重复使用buf的前提是这里传值而不是传指针

		// 交给router处理
		go func(req *Request) {
			for _, ri := range sc.Routers {
				ri.PreHandler(req)
				ri.Handler(req)
				ri.AfterHandler(req)
			}
		}(req)
	}
	// fmt.Println("close conn") // 客户端退出，触发超时，关闭流
	sc.Close()
}

func (sc *ServerConnection) bindRouters(rs []RouterI) {
	sc.Routers = append(sc.Routers, rs...)
}

// Client Connection impliment
type ConnectionIC interface {
	ConnectionI
	// for Client connection
	OpenNewDataStream() (quic.Stream, StreamID, error)
}

type ClientConnection struct {
	*Connection
}

func NewClientConnection(ctx context.Context, qconn quic.Connection) *ClientConnection {
	conn := newConnection(ctx)

	// init other value
	conn.conn = qconn

	return &ClientConnection{Connection: &conn}
}

func (cc *ClientConnection) OpenNewDataStream() (quic.Stream, StreamID, error) {
	if cc.controlStream == nil {
		return nil, 0, fmt.Errorf("controlStream is nil")
	}

	cc.controlStream.WriteFrame(frame.NewRequestDataStreamFrame())

	fmt.Println("request new data stream, awaiting")
	newstream, err := cc.conn.AcceptStream(cc.ctx)
	if err != nil {
		return nil, 0, err
	}
	fmt.Println("reqeust success, new data stream", newstream.StreamID())
	cc.streams.Store(newstream.StreamID(), newstream)

	return newstream, StreamID(newstream.StreamID()), nil
}
