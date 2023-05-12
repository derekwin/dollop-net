package dollop

import (
	"errors"

	"github.com/quic-go/quic-go"
)

// ErrStreamNil be returned if FrameStream underlying stream is nil.
var ErrControlStreamNil = errors.New("control stream underlying is nil")

// ControlStreamI based FrameStream with control related parse function
type ControlStreamI interface {
	FrameStreamI
	ParseMsg(frame *Frame) ControlMsgI // input a &Frame{}, output a &ControlMsg
}

// ControlStream is the frame.ReadWriter that goroutinue read write safely.
type ControlStream struct {
	FrameStream
	msgParser ParseControlMsgHandler // message paser
}

// NewControlStream creates a new ControlStream.
func NewControlStream(s quic.Stream) *ControlStream {
	return &ControlStream{FrameStream: FrameStream{stream: s}, msgParser: ParseControlMsg} // use default contril msg parser
}

func (cs *ControlStream) ParseMsg(frame *Frame) ControlMsgI {
	return cs.msgParser(frame)
}
