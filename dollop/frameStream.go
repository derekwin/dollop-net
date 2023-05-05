package dollop

import (
	"errors"
	"sync"

	"github.com/derekwin/dollop-net/dollop/frame"
	"github.com/quic-go/quic-go"
)

// ErrStreamNil be returned if FrameStream underlying stream is nil.
var ErrStreamNil = errors.New("frame stream underlying is nil")

type FrameStreamI interface {
	StreamID() StreamID
	ReadFrame() (frame.Frame, error)
	WriteFrame(f frame.Frame) error
	Close()
}

// FrameStream is the frame.ReadWriter that goroutinue read write safely.
type FrameStream struct {
	stream quic.Stream
	mu     sync.Mutex
}

// NewFrameStream creates a new FrameStream.
func NewFrameStream(s quic.Stream) FrameStreamI {
	return &FrameStream{stream: s}
}

func (fs *FrameStream) StreamID() StreamID {
	return StreamID(fs.stream.StreamID())
}

// ReadFrame reads next frame from underlying stream.
func (fs *FrameStream) ReadFrame() (frame.Frame, error) {
	if fs.stream == nil {
		return nil, ErrStreamNil
	}
	return frame.ParseFrame(fs.stream)
}

// WriteFrame writes a frame into underlying stream.
func (fs *FrameStream) WriteFrame(f frame.Frame) error {
	if fs.stream == nil {
		return ErrStreamNil
	}
	fs.mu.Lock()
	defer fs.mu.Unlock()

	_, err := fs.stream.Write(f.Encode())
	return err
}

func (fs *FrameStream) Close() {
	fs.stream.Close()
}
