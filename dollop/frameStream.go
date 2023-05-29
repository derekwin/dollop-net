package dollop

import (
	"encoding/binary"
	"errors"
	"io"
	"sync"

	"github.com/quic-go/quic-go"
)

// ErrStreamNil be returned if FrameStream underlying stream is nil.
var ErrFrameStreamNil = errors.New("FrameStream's stream is nil")

type FrameStreamI interface {
	StreamID() StreamID
	readFrame() (*Frame, error)
	writeFrame(f *Frame) error
	BindMsgProtocol(msgP MsgProtocolI) // 协议绑定机制，将协议绑定到帧流上；子流级别增加新协议支持
	GetRouter(tag MsgType) (FrameRouterI, error)
	ReadMsg() (MsgI, error) // 根据绑定的消息协议，完成帧到msg一步到位解析
	WriteMsg(m MsgI) error  // 根据绑定的消息协议，将msg包装成帧发送
	Close()
}

// FrameStream is the ReadWriter that goroutinue read write safely.
type FrameStream struct {
	stream      quic.Stream
	msgProtocol MsgProtocolI
	mu          sync.Mutex
}

// NewFrameStream creates a new FrameStream.
func NewFrameStream(s quic.Stream) *FrameStream {
	return &FrameStream{stream: s}
}

func (fs *FrameStream) StreamID() StreamID {
	return StreamID(fs.stream.StreamID())
}

// Frame :  | len:FrameLen | Msg |
func readFrame(stream io.Reader) (*Frame, error) {
	lenBuf := make([]byte, FrameLen)
	_, err := stream.Read(lenBuf)
	if err != nil {
		return &Frame{}, err
	}

	bufferLen := binary.BigEndian.Uint32(lenBuf)

	frameBuf := make([]byte, bufferLen)
	_, err = stream.Read(frameBuf)
	if err != nil {
		return &Frame{}, err
	}
	//  Frame
	return NewFrame(frameBuf), nil
}

// ReadFrame reads next frame from underlying stream.
func (fs *FrameStream) readFrame() (*Frame, error) {
	if fs.stream == nil {
		return &Frame{}, ErrFrameStreamNil
	}
	return readFrame(fs.stream)
}

// WriteFrame writes a frame into underlying stream.
func (fs *FrameStream) writeFrame(f *Frame) error {
	if fs.stream == nil {
		return ErrFrameStreamNil
	}
	fs.mu.Lock()
	defer fs.mu.Unlock()

	_, err := fs.stream.Write(f.Encode())
	return err
}

func (fs *FrameStream) Close() {
	fs.stream.Close()
}

func (fs *FrameStream) BindMsgProtocol(msgP MsgProtocolI) {
	fs.msgProtocol = msgP
}

func (fs *FrameStream) GetRouter(tag MsgType) (FrameRouterI, error) {
	return fs.msgProtocol.GetRouter(tag)
}

func (fs *FrameStream) ReadMsg() (MsgI, error) {
	f, err := fs.readFrame()
	if err != nil {
		return nil, err
	}

	return fs.msgProtocol.PaserMsg(f), nil
}

func (fs *FrameStream) WriteMsg(m MsgI) error {
	f := NewFrame(m.Encode())
	fs.writeFrame(f)
	return nil
}
