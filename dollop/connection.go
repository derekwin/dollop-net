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
	GetRawStream(id StreamID) (RawStreamI, error)
	GetFrameStream(id StreamID) (FrameStreamI, error)
	setControlStream(stream FrameStreamI)
	addRawStream(id StreamID, stream RawStreamI)
	addFrameStream(id StreamID, stream FrameStreamI)
	deleteRawStream(id StreamID)
	deleteFrameStream(id StreamID)
	Close()
}

// Connection 初始启动时候，建立一条controlStream，类型是FrameStream，用于管理后续流的建立和维护。提供两种流，不进行分包的流quic.stream, 分帧的流 FrameStream
type Connection struct {
	ctx                    context.Context
	conn                   quic.Connection
	controlStream          FrameStreamI                        // control stream
	requestRawStreamChan   chan *frame.RequestRawStreamFrame   // 管理无分包的流
	requestFrameStreamChan chan *frame.RequestFrameStreamFrame // 管理帧流
	rawStreams             sync.Map                            // 无分包的流 RawStreamI
	frameStreams           sync.Map                            // 帧流 FrameStreamI use RawFrame
	group                  sync.WaitGroup
}

func newConnection(ctx context.Context) Connection {
	return Connection{ctx: ctx, requestRawStreamChan: make(chan *frame.RequestRawStreamFrame, 10),
		requestFrameStreamChan: make(chan *frame.RequestFrameStreamFrame, 10)}
}

func (c *Connection) Close() {

	c.controlStream.Close()

	c.rawStreams.Range(func(key, value interface{}) bool {
		stream := value.(RawStreamI)
		stream.Close()
		return true
	})

	c.frameStreams.Range(func(key, value interface{}) bool {
		stream := value.(FrameStreamI)
		stream.Close()
		return true
	})

	c.conn.CloseWithError(ServerConnectionCloseCode, "server side close this conn")
}

func (c *Connection) GetConn() quic.Connection {
	return c.conn
}

func (c *Connection) GetRawStream(id StreamID) (RawStreamI, error) {
	stream, ok := c.rawStreams.Load(id)
	if !ok {
		return nil, fmt.Errorf("err get valid stream")
	}
	return stream.(RawStreamI), nil
}

func (c *Connection) GetFrameStream(id StreamID) (FrameStreamI, error) {
	stream, ok := c.frameStreams.Load(id)
	if !ok {
		return &FrameStream{}, fmt.Errorf("err get valid stream")
	}
	return stream.(FrameStreamI), nil
}

func (c *Connection) setControlStream(stream FrameStreamI) {
	c.controlStream = stream
}

func (c *Connection) addRawStream(id StreamID, stream RawStreamI) {
	c.rawStreams.Store(id, stream)
}

func (c *Connection) addFrameStream(id StreamID, stream FrameStreamI) {
	c.frameStreams.Store(id, stream)
}

func (c *Connection) deleteRawStream(id StreamID) {
	c.rawStreams.Delete(id)
}

func (c *Connection) deleteFrameStream(id StreamID) {
	c.frameStreams.Delete(id)
}

func (c *Connection) Wait() { c.group.Wait() }

// Server Connection impliment specisal
type ConnectionIS interface {
	ConnectionI
	// for Server connection
	Serve()
	bindRawRouters([]RawRouterI)
	bindFrameRouters([]FrameRouterI)
	controlStreamLoop()
}

type ServerConnection struct {
	Connection
	RawRouters   []RawRouterI
	FrameRouters []FrameRouterI
}

func NewServerConnection(ctx context.Context, qconn quic.Connection) *ServerConnection {
	conn := newConnection(ctx)

	// init other value
	conn.conn = qconn

	return &ServerConnection{Connection: conn}
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
		case frame.RequestRawStreamFrame:
			sc.requestRawStreamChan <- &ff
		case frame.RequestFrameStreamFrame:
			sc.requestFrameStreamChan <- &ff
		default:
			fmt.Println("control stream read unexcepted frame", "frame_type", f.Type())
		}
	}
}

func (sc *ServerConnection) streamManager(ctx context.Context) chan struct{} {
	sc.group.Add(2)
	go sc.dataStreamManager(ctx)
	go sc.frameStreamManager(ctx)
	sc.group.Wait()
	return make(chan struct{})
}

func (sc *ServerConnection) dataStreamManager(ctx context.Context) chan struct{} {
	for {
		ff, ok := <-sc.requestRawStreamChan
		if !ok {
			fmt.Println(io.EOF)
		}

		// err := handshake(ff)

		// 如果出错，返回拒绝
		// cs.controlStream.WriteFrame(frame.NewRejectRawStreamFrame())

		fmt.Println("receive new data stream request", ff)
		newStream, err := sc.conn.OpenStreamSync(ctx)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("open new data stream", newStream.StreamID())
		_, err = newStream.Write(frame.NewAckStreamFrame().Encode())
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("ack to new data stream")

		sc.addRawStream(StreamID(newStream.StreamID()), newStream)

		go sc.processRawStream(newStream)
	}
}

func (sc *ServerConnection) frameStreamManager(ctx context.Context) chan struct{} {
	for {
		ff, ok := <-sc.requestFrameStreamChan
		if !ok {
			fmt.Println(io.EOF)
		}

		// err := handshake(ff)

		// 如果出错，返回拒绝
		// cs.controlStream.WriteFrame(frame.NewRejectRawStreamFrame())

		fmt.Println("receive new frame stream request", ff)
		newQuicStream, err := sc.conn.OpenStreamSync(ctx)
		if err != nil {
			fmt.Println(err)
		}
		newStream := NewFrameStream(newQuicStream)
		fmt.Println("open new frame stream", newQuicStream.StreamID())
		err = newStream.WriteFrame(frame.NewAckStreamFrame())
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("ack to new frame stream")

		sc.addFrameStream(StreamID(newQuicStream.StreamID()), newStream)

		go sc.processFrameStream(newStream)
	}
}

func (sc *ServerConnection) processRawStream(stream RawStreamI) {
	fmt.Println("process data stream", stream.StreamID())
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
		req := &RawRequest{conn: sc, stream: stream, data: buf} // 重复使用buf的前提是这里传值而不是传指针

		// 交给router处理
		go func(req *RawRequest) {
			for _, ri := range sc.RawRouters {
				ri.PreHandler(req)
				ri.Handler(req)
				ri.AfterHandler(req)
			}
		}(req)
	}
	// fmt.Println("close conn") // 客户端退出，触发超时，关闭流
	sc.Close()
}

func (sc *ServerConnection) processFrameStream(stream FrameStreamI) {
	fmt.Println("process frame stream", stream.StreamID())
	for {
		// 判断ctx业务退出? 是否有必要

		// 读取数据
		f, err := stream.ReadFrame()
		if err != nil {
			// fmt.Println(err) // 客户端退出后，会触发超时
			break
		}
		// 将数据请求封装为request，然后分别调用对应的router
		// 生成request
		req := &FrameRequest{conn: sc, stream: stream, data: f.GetData()} // 重复使用buf的前提是这里传值而不是传指针

		// 交给router处理
		go func(req *FrameRequest) {
			for _, ri := range sc.FrameRouters {
				ri.PreHandler(req)
				ri.Handler(req)
				ri.AfterHandler(req)
			}
		}(req)
	}
	// fmt.Println("close conn") // 客户端退出，触发超时，关闭流
	sc.Close()
}

func (sc *ServerConnection) bindRawRouters(rs []RawRouterI) {
	sc.RawRouters = append(sc.RawRouters, rs...)
}

func (sc *ServerConnection) bindFrameRouters(rs []FrameRouterI) {
	sc.FrameRouters = append(sc.FrameRouters, rs...)
}

// Client Connection impliment
type ConnectionIC interface {
	ConnectionI
	// for Client connection
	OpenNewRawStream() (RawStreamI, StreamID, error)
}

type ClientConnection struct {
	Connection
}

func NewClientConnection(ctx context.Context, qconn quic.Connection) *ClientConnection {
	conn := newConnection(ctx)

	// init other value
	conn.conn = qconn

	return &ClientConnection{Connection: conn}
}

func (cc *ClientConnection) OpenNewRawStream() (RawStreamI, StreamID, error) {
	if cc.controlStream == nil {
		return nil, 0, fmt.Errorf("controlStream is nil")
	}

	cc.controlStream.WriteFrame(frame.NewRequestRawStreamFrame())

	fmt.Println("request new data stream, awaiting")
	newstream, err := cc.conn.AcceptStream(cc.ctx)
	if err != nil {
		return nil, 0, err
	}
	fmt.Println("reqeust success, new data stream", newstream.StreamID())
	cc.rawStreams.Store(newstream.StreamID(), newstream)

	return newstream, StreamID(newstream.StreamID()), nil
}

func (cc *ClientConnection) OpenNewFrameStream() (FrameStreamI, StreamID, error) {
	if cc.controlStream == nil {
		return nil, 0, fmt.Errorf("controlStream is nil")
	}

	cc.controlStream.WriteFrame(frame.NewRequestFrameStreamFrame())

	fmt.Println("request new frame stream, awaiting")
	newQuicstream, err := cc.conn.AcceptStream(cc.ctx)
	if err != nil {
		return nil, 0, err
	}
	fmt.Println("get quic stream")
	newStream := NewFrameStream(newQuicstream)
	f, err := newStream.ReadFrame()
	if err != nil {
		return nil, 0, err
	}

	switch f.(type) {
	case frame.AckStreamFrame:
		fmt.Println("reqeust success, new frame stream", newQuicstream.StreamID())
		cc.frameStreams.Store(newQuicstream.StreamID(), newStream)
		return newStream, StreamID(newQuicstream.StreamID()), nil
	default:
		return nil, 0, fmt.Errorf("not receive Ack")
	}

}
