package dollop

import (
	"bytes"
	"encoding/binary"
)

// Type represents the type of frame.
const FrameLen int = 4 // int32

// 帧在本框架是固定的存在，帧流的最小单元永远是Frame
type Frame struct {
	len  int // how long this frame
	data []byte
}

func (f Frame) GetData() []byte {
	return f.data
}

func (f Frame) Encode() []byte {
	frameBuf := bytes.NewBuffer([]byte{})

	binary.Write(frameBuf, binary.BigEndian, uint32(f.len))
	binary.Write(frameBuf, binary.BigEndian, f.data)

	//  | len [FrameLen] | Msg |
	return frameBuf.Bytes()
}

func NewFrame(data []byte) *Frame {
	return &Frame{len: len(data), data: data}
}
