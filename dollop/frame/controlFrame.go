package frame

import (
	"bytes"
	"encoding/binary"
)

// client send RequestDataSreamFrame to apply a new stream from server
type RequestDataStreamFrame struct {
	streamId StreamID
}

func (rdsf RequestDataStreamFrame) Type() Type {
	return RequestDataStreamFrameTag
}

func (rdsf RequestDataStreamFrame) StreamId() StreamID {
	return rdsf.streamId
}

func (rdsf RequestDataStreamFrame) Encode() []byte {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.BigEndian, RequestDataStreamFrameTag)
	return buf.Bytes()
}

func NewRequestDataStreamFrame(id StreamID) RequestDataStreamFrame {
	return RequestDataStreamFrame{streamId: id}
}

func ParseRequestDataStreamFrame(buf []byte) (RequestDataStreamFrame, error) {
	return RequestDataStreamFrame{}, nil
}

// AckDataStreamFrame sent from server to client after client sent RequestDataSreamFrame
type AckDataStreamFrame struct {
	streamId StreamID
}

func (adsf AckDataStreamFrame) Type() Type {
	return AckDataStreamFrameTag
}

func (adsf AckDataStreamFrame) Encode() []byte {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.BigEndian, AckDataStreamFrameTag)
	return buf.Bytes()
}

func NewAckDataStreamFrame(id StreamID) AckDataStreamFrame {
	return AckDataStreamFrame{streamId: id}
}

func ParseAckDataStreamFrame(buf []byte) (AckDataStreamFrame, error) {
	return AckDataStreamFrame{}, nil
}

// RejectDataStreamFrame sent from server to client while occur err
type RejectDataStreamFrame struct {
	streamId StreamID
}

func (adsf RejectDataStreamFrame) Type() Type {
	return RejectDataStreamFrameTag
}

func (adsf RejectDataStreamFrame) Encode() []byte {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.BigEndian, RejectDataStreamFrameTag)
	return buf.Bytes()
}

func NewRejectDataStreamFrame(id StreamID) RejectDataStreamFrame {
	return RejectDataStreamFrame{streamId: id}
}

func ParseRejectDataStreamFrame(buf []byte) (RejectDataStreamFrame, error) {
	return RejectDataStreamFrame{}, nil
}
