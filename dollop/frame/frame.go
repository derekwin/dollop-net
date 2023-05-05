package frame

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// Type represents the type of frame.
type Type uint8

const TagLen int = 1   // uint8
const FrameLen int = 4 // int32

const (
	// data Frame Tag
	DataFrameTag Type = 0x01

	// control Frame Tag
	RequestRawStreamFrameTag   Type = 0x02
	RequestFrameStreamFrameTag Type = 0x03
	AckStreamFrameTag          Type = 0x04
	RejectStreamFrameTag       Type = 0x05
)

type Frame interface {
	// Type gets the type of Frame.
	Type() Type

	// Encode the frame into []byte.
	Encode() []byte

	GetData() []byte
}

// read raw frame from quic.stream
func ReadFrame(stream io.Reader) ([]byte, error) {
	//   | len [FrameLen] | type [Type] | data |
	lenBuf := make([]byte, FrameLen)
	_, err := stream.Read(lenBuf)
	if err != nil {
		return nil, err
	}

	len := binary.BigEndian.Uint32(lenBuf)
	// fmt.Println(len, FrameLen, lenBuf)

	frameBuf := make([]byte, len+uint32(TagLen))
	_, err = stream.Read(frameBuf)
	if err != nil {
		return nil, err
	}
	//   | type [Type] | data |
	return frameBuf, nil
}

// write raw frame with len
func WriteFrame(data []byte, frameType Type) []byte {
	//   | len [FrameLen] | type [Type] | data |
	frameLen := len(data) + TagLen
	frameBuf := bytes.NewBuffer([]byte{})

	binary.Write(frameBuf, binary.BigEndian, uint32(frameLen))
	binary.Write(frameBuf, binary.BigEndian, frameType)
	binary.Write(frameBuf, binary.BigEndian, data)

	return frameBuf.Bytes()
}

// ParseFrame parses the frame from QUIC stream.
func ParseFrame(stream io.Reader) (Frame, error) {
	// buf, err := y3.ReadPacket(stream)
	// if err != nil {
	// 	return nil, err
	// }
	buf, err := ReadFrame(stream)
	if err != nil {
		return nil, err
	}

	frameType := buf[0]

	switch frameType {
	case byte(RequestRawStreamFrameTag):
		return ParseRequestRawStreamFrame(buf)
	case byte(AckStreamFrameTag):
		return ParseAckStreamFrame(buf)
	case byte(RejectStreamFrameTag):
		return ParseRejectStreamFrame(buf)
	case byte(DataFrameTag):
		return ParseDataFrame(buf)
	case byte(RequestFrameStreamFrameTag):
		return ParseRequestFrameStreamFrame(buf)
	default:
		return nil, fmt.Errorf("unknown frame type, buf[0]=%#x", buf[0])
	}
}
