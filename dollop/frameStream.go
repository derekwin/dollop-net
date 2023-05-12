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
	ReadFrame() (*Frame, error)
	WriteFrame(f *Frame) error
	Close()
}

// FrameStream is the ReadWriter that goroutinue read write safely.
type FrameStream struct {
	stream quic.Stream
	mu     sync.Mutex
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
func (fs *FrameStream) ReadFrame() (*Frame, error) {
	if fs.stream == nil {
		return &Frame{}, ErrFrameStreamNil
	}
	return readFrame(fs.stream)
}

// WriteFrame writes a frame into underlying stream.
func (fs *FrameStream) WriteFrame(f *Frame) error {
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
