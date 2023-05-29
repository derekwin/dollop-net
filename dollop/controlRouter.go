package dollop

import "fmt"

type RequestFrameStreamRouter struct {
	BaseFrameRouter
}

func (rfsr RequestFrameStreamRouter) Handler(req FrameRequestI) error {
	conn, err := req.GetConn()
	if err != nil {
		return err
	}
	stream, err := req.GetStream()
	if err != nil {
		return err
	}
	data, err := req.GetData()
	if err != nil {
		return err
	}
	fmt.Printf("RequestFrameStreamRouterHandler : stream: %d, request data:%s", stream.StreamID(), data)

	// 如果出错，返回拒绝
	// stream.WriteMsg(NewRejectRawStreamMsg())
	newQuicStream, err := conn.OpenStreamSync()
	if err != nil {
		fmt.Println(err)
	}
	newStream := NewFrameStream(newQuicStream)
	fmt.Println("open new frame stream", newQuicStream.StreamID())

	newStream.BindMsgProtocol(defaultMsgProtocol)

	err = newStream.WriteMsg(NewAckStreamMsg([]byte{}))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("ack to the new frame stream")

	conn.addFrameStream(StreamID(newQuicStream.StreamID()), newStream)

	return nil
}

type RequestRawStreamRouter struct {
	BaseFrameRouter
}

func (rrsr RequestRawStreamRouter) Handler(req FrameRequestI) error {
	conn, err := req.GetConn()
	if err != nil {
		return err
	}
	sconn := conn.(ConnectionIS) // 实际为server流
	stream, err := req.GetStream()
	if err != nil {
		return err
	}
	data, err := req.GetData()
	if err != nil {
		return err
	}
	fmt.Printf("RequestRawStreamRouterHandler : stream: %d, request data:%s", stream.StreamID(), data)

	// 如果出错，返回拒绝
	// stream.WriteMsg(NewRejectRawStreamMsg())
	newQuicStream, err := sconn.OpenStreamSync()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("open new raw stream", newQuicStream.StreamID())

	newStream := NewRawStream(newQuicStream)
	newStream.BindRawRouter(&BaseRawRouter{})

	_, err = newStream.Write(NewAckStreamMsg([]byte{}).Encode())
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("ack to the new raw stream")

	conn.addRawStream(StreamID(newQuicStream.StreamID()), newStream)

	go sconn.ProcessRawStream(newStream)
	return nil
}
