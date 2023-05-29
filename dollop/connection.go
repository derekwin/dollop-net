package dollop

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/quic-go/quic-go"
)

type StreamID int64

const ServerConnectionCloseCode quic.ApplicationErrorCode = 701

type ConnectionI interface {
	GetQConn() (quic.Connection, error)
	GetRawStream(id StreamID) (RawStreamI, error)          // *RawStream
	GetFrameStream(id StreamID) (FrameStreamI, error)      // *FrameStream
	setControlStream(stream FrameStreamI) error            // *ControlStream
	addRawStream(id StreamID, stream RawStreamI) error     // *RawStrea
	addFrameStream(id StreamID, stream FrameStreamI) error // *FrameStream
	deleteRawStream(id StreamID) error
	deleteFrameStream(id StreamID) error
	OpenStreamSync() (quic.Stream, error)
	Close() error
}

// Connection 初始启动时候，建立一条controlStream，类型是FrameStream，用于管理后续流的建立和维护。提供两种流，不进行分包的流quic.stream, 分帧的流 FrameStream
type Connection struct {
	ctx           context.Context
	qconn         quic.Connection
	controlStream FrameStreamI // *ControlStream
	rawStreams    sync.Map     // 无分包的流 RawStreamI *RawStream
	frameStreams  sync.Map     // 帧流 FrameStreamI *FrameStream
	group         sync.WaitGroup
}

func (c *Connection) Close() error {

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

	return c.qconn.CloseWithError(ServerConnectionCloseCode, "server side close this conn")
}

func (c *Connection) GetQConn() (quic.Connection, error) {
	return c.qconn, nil
}

func (c *Connection) GetRawStream(id StreamID) (RawStreamI, error) {
	stream, ok := c.rawStreams.Load(id)
	if !ok {
		return nil, fmt.Errorf("err get valid raw stream")
	}
	return stream.(RawStreamI), nil
}

func (c *Connection) GetFrameStream(id StreamID) (FrameStreamI, error) {
	stream, ok := c.frameStreams.Load(id)
	if !ok {
		return nil, fmt.Errorf("err get valid frame stream")
	}
	return stream.(FrameStreamI), nil
}

func (c *Connection) setControlStream(stream FrameStreamI) error {
	c.controlStream = stream
	return nil
}

func (c *Connection) addRawStream(id StreamID, stream RawStreamI) error {
	c.rawStreams.Store(id, stream)
	return nil
}

func (c *Connection) addFrameStream(id StreamID, stream FrameStreamI) error {
	c.frameStreams.Store(id, stream)
	return nil
}

func (c *Connection) deleteRawStream(id StreamID) error {
	c.rawStreams.Delete(id)
	return nil
}

func (c *Connection) deleteFrameStream(id StreamID) error {
	c.frameStreams.Delete(id)
	return nil
}

func (c *Connection) OpenStreamSync() (quic.Stream, error) {
	return c.qconn.OpenStreamSync(c.ctx)
}

func (c *Connection) Wait() { c.group.Wait() }

// Server Connection impliment specisal
type ConnectionIS interface {
	ConnectionI
	// for Server connection
	Serve(ctx context.Context) <-chan struct{}
	BindRawRouters([]RawRouterI)
	// 绑定frame流对应的协议
	BindMsgProtocol(sId StreamID, mP MsgProtocolI) error
	controlStreamLoop()
	ProcessRawStream(stream RawStreamI)
	ProcessFrameStream(stream FrameStreamI)
}

type ServerConnection struct {
	Connection
	RawRouters                []RawRouterI
	requestRawStreamMsgChan   chan *RequestRawStreamMsg   // 管理无分包的流
	requestFrameStreamMsgChan chan *RequestFrameStreamMsg // 管理分包的流
	// FrameRouters []FrameRouterI
}

func NewServerConnection(ctx context.Context, qconn quic.Connection) *ServerConnection {
	return &ServerConnection{Connection: Connection{ctx: ctx, qconn: qconn},
		requestRawStreamMsgChan: make(chan *RequestRawStreamMsg, 10), requestFrameStreamMsgChan: make(chan *RequestFrameStreamMsg, 10)}
}

func (sc *ServerConnection) Serve(ctx context.Context) <-chan struct{} {
	done := make(chan struct{})

	qconn := sc.qconn

	// wg := new(sync.WaitGroup)

	qStream, err := qconn.AcceptStream(ctx)
	if err != nil {
		fmt.Println(err)
	}
	defer qStream.Close()

	controlStream := NewFrameStream(qStream)
	controlStream.BindMsgProtocol(controlMsgProtocol)

	sc.setControlStream(controlStream)

	// 启动流管理器
	go sc.controlStreamLoop()

	// 进行控制流管理循环
	go func(sc *ServerConnection) {

		defer sc.Close()

		select {
		case <-ctx.Done():
			return
		case <-sc.StreamManager(ctx):
		}
	}(sc)

	return done
}

func (sc *ServerConnection) controlStreamLoop() {
	// if sc.requestRawStreamMsgChan == nil {
	// 	sc.requestRawStreamMsgChan = make(chan MsgI, 10)
	// }
	// if sc.requestFrameStreamMsgChan == nil {
	// 	sc.requestFrameStreamMsgChan = make(chan MsgI, 10)
	// }

	for {
		m, err := sc.controlStream.ReadMsg()
		if err != nil {
			sc.qconn.CloseWithError(0, err.Error())
			return
		}

		// fmt.Println(thisMsgType.)
		fmt.Println(sc.requestRawStreamMsgChan)
		typeCode := m.Type()
		switch typeCode.(ControlMsgType) { // TODO, 这里确实需要msg.(type)才能变到这样，怀疑是多层ControlMsgI返回导致无法解析到类型,做到协议，只做一层传出
		case RequestRawStreamMsgTag:
			fmt.Println("tes1")
			sc.requestRawStreamMsgChan <- m.(*RequestRawStreamMsg)
			fmt.Println("tes1")
		case RequestFrameStreamMsgTag:
			fmt.Println("tes2")
			sc.requestFrameStreamMsgChan <- m.(*RequestFrameStreamMsg)
			fmt.Println("tes2")
		default:
			fmt.Println("default", m)
			fmt.Println("control stream read unexcepted", "control msg type")
		}
	}
}

func (sc *ServerConnection) StreamManager(ctx context.Context) chan struct{} {
	// sc.group.Add(2)
	go sc.rawStreamManager(ctx)
	fmt.Println("start raw stream manager")
	go sc.frameStreamManager(ctx)
	fmt.Println("start frame stream manager")
	return make(chan struct{})
}

func (sc *ServerConnection) rawStreamManager(ctx context.Context) chan struct{} {
	for {
		fmt.Printf("awating a new raw msg\n")

		ff, ok := <-sc.requestRawStreamMsgChan
		fmt.Println(ff)
		if !ok {
			fmt.Println(io.EOF)
		}
		msg := ff
		// 将msg转为request，此举是为了后续扩展多个msg形成一个request
		req := &FrameRequest{conn: sc, stream: sc.controlStream, msg: msg}

		router, err := sc.controlStream.GetRouter(msg.Type())
		if err != nil {
			fmt.Println(err)
		}

		go func(req *FrameRequest) {
			router.PreHandler(req)
			router.Handler(req)
			router.AfterHandler(req)
		}(req)
	}
}

func (sc *ServerConnection) frameStreamManager(ctx context.Context) chan struct{} {
	for {
		fmt.Printf("awating a new frame msg")
		ff, ok := <-sc.requestFrameStreamMsgChan
		fmt.Println(ff)
		if !ok {
			fmt.Println(io.EOF)
		}
		msg := ff
		// 将msg转为request，此举是为了后续扩展多个msg形成一个request
		req := &FrameRequest{conn: sc, stream: sc.controlStream, msg: msg}

		router, err := sc.controlStream.GetRouter(msg.Type())
		if err != nil {
			fmt.Println(err)
		}

		go func(req *FrameRequest) {
			router.PreHandler(req)
			router.Handler(req)
			router.AfterHandler(req)
		}(req)
	}
}

func (sc *ServerConnection) ProcessRawStream(stream RawStreamI) {
	fmt.Println("process raw stream", stream.StreamID())
	buf := make([]byte, 512) // 分配一次，重复使用 // TODO，将切分逻辑交给路由
	for {
		// 判断ctx业务退出?

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

func (sc *ServerConnection) ProcessFrameStream(stream FrameStreamI) {
	fmt.Println("process frame stream", stream.StreamID())
	for {
		// 判断ctx业务退出? 是否有必要

		// 读取数据
		f, err := stream.ReadMsg()
		if err != nil {
			// fmt.Println(err) // 客户端退出后，会触发超时
			break
		}
		// 将数据请求封装为request，然后分别调用对应的router
		// 生成request
		req := &FrameRequest{conn: sc, stream: stream, msg: f}
		router, err := stream.GetRouter(f.Type())
		if err != nil {
			fmt.Println(err)
		}

		// 交给router处理
		go func(req *FrameRequest) {
			router.PreHandler(req)
			router.Handler(req)
			router.AfterHandler(req)
		}(req)
	}
	// fmt.Println("close conn") // 客户端退出，触发超时，关闭流
	sc.Close()
}

func (sc *ServerConnection) BindRawRouters(rs []RawRouterI) {
	sc.RawRouters = append(sc.RawRouters, rs...)
}

func (sc *ServerConnection) BindMsgProtocol(sId StreamID, mP MsgProtocolI) error {
	stream, err := sc.GetFrameStream(sId)
	if err != nil {
		return err
	}

	stream.BindMsgProtocol(mP)
	return nil
}

// Client Connection impliment
type ConnectionIC interface {
	ConnectionI
	// for Client connection
	OpenNewRawStream() (RawStreamI, StreamID, error)
	OpenNewFrameStream() (FrameStreamI, StreamID, error)
}

type ClientConnection struct {
	Connection
}

func NewClientConnection(ctx context.Context, qconn quic.Connection) *ClientConnection {
	return &ClientConnection{Connection: Connection{ctx: ctx, qconn: qconn}}
}

func (cc *ClientConnection) OpenNewRawStream() (RawStreamI, StreamID, error) {
	if cc.controlStream == nil {
		return nil, 0, fmt.Errorf("controlStream is nil")
	}

	cc.controlStream.WriteMsg(NewRequestRawStreamMsg([]byte{}))

	fmt.Println("request new raw stream, awaiting")
	newQStream, err := cc.qconn.AcceptStream(cc.ctx)
	if err != nil {
		return nil, 0, err
	}
	fmt.Println("reqeust success, new data stream", newQStream.StreamID())
	newStream := NewRawStream(newQStream)

	cc.rawStreams.Store(newStream.StreamID(), newStream)

	return newStream, StreamID(newStream.StreamID()), nil
}

func (cc *ClientConnection) OpenNewFrameStream() (FrameStreamI, StreamID, error) {
	if cc.controlStream == nil {
		return nil, 0, fmt.Errorf("controlStream is nil")
	}

	cc.controlStream.WriteMsg(NewRequestFrameStreamMsg([]byte{}))

	fmt.Println("request new frame stream, awaiting")
	newQStream, err := cc.qconn.AcceptStream(cc.ctx)
	if err != nil {
		return nil, 0, err
	}
	fmt.Println("get quic stream")
	newStream := NewFrameStream(newQStream)
	f, err := newStream.ReadMsg()
	if err != nil {
		return nil, 0, err
	}

	switch f.(type) {
	case AckStreamMsg:
		fmt.Println("reqeust success, new frame stream", newStream.StreamID())
		cc.frameStreams.Store(newStream.StreamID(), newStream)
		return newStream, StreamID(newStream.StreamID()), nil
	default:
		return nil, 0, fmt.Errorf("not receive Ack")
	}
}
