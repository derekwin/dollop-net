package frame

import (
	"fmt"
	"io"
)

type StreamID int64

// Type represents the type of frame.
type Type uint8

const (
	RequestDataStreamFrameTag Type = 0x01
	AckDataStreamFrameTag     Type = 0x02
	RejectDataStreamFrameTag  Type = 0x03
)

type Frame interface {
	// Type gets the type of Frame.
	Type() Type

	// Encode the frame into []byte.
	Encode() []byte
}

// ParseFrame parses the frame from QUIC stream.
func ParseFrame(stream io.Reader) (Frame, error) {
	// buf, err := y3.ReadPacket(stream)
	// if err != nil {
	// 	return nil, err
	// }
	buf := make([]byte, 512)
	_, err := stream.Read(buf)
	if err != nil {
		return nil, err
	}

	frameType := buf[0]
	fmt.Println(frameType)
	// fmt.Println(0x80 | byte(RequestDataStreamFrameTag))
	// fmt.Println(0x80 | byte(AckDataStreamFrameTag))
	switch frameType {
	// case 0x80 | byte(RequestDataStreamFrameTag):
	// 	return ParseRequestDataStreamFrame(buf)
	// case 0x80 | byte(AckDataStreamFrameTag):
	// 	return ParseAckDataStreamFrame(buf)
	case byte(RequestDataStreamFrameTag):
		return ParseRequestDataStreamFrame(buf)
	case byte(AckDataStreamFrameTag):
		return ParseAckDataStreamFrame(buf)
	default:
		return nil, fmt.Errorf("unknown frame type, buf[0]=%#x", buf[0])
	}
}
