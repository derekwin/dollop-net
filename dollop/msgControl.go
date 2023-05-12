package dollop

import (
	"bytes"
	"encoding/binary"
)

type Type uint8

const TagLen int = 1 // Type's byte len : uint8 -> 1

const (
	// control Frame Tag
	RequestRawStreamMsgTag   Type = 0x01
	RequestFrameStreamMsgTag Type = 0x02
	AckStreamMsgTag          Type = 0x03
	RejectStreamMsgTag       Type = 0x04
)

type ControlMsgI interface {
	MsgI
}

func NewControlMsg(msgTag Type, data []byte) ControlMsgI {
	switch msgTag {
	case RequestRawStreamMsgTag:
		return &RequestRawStreamMsg{data: data}
	case RequestFrameStreamMsgTag:
		return &RequestFrameStreamMsg{data: data}
	case AckStreamMsgTag:
		return &AckStreamMsg{data: data}
	case RejectStreamMsgTag:
		return &RejectStreamMsg{data: data}
	}
	return nil
}

type ParseControlMsgHandler func(*Frame) ControlMsgI

func ParseControlMsg(frame *Frame) ControlMsgI {

	msgType := frame.data[0]
	dataBuf := frame.data[TagLen:]

	switch msgType {
	case byte(RequestRawStreamMsgTag):
		return NewControlMsg(RequestRawStreamMsgTag, dataBuf)
	case byte(RequestFrameStreamMsgTag):
		return NewControlMsg(RequestFrameStreamMsgTag, dataBuf)
	case byte(AckStreamMsgTag):
		return NewControlMsg(AckStreamMsgTag, dataBuf)
	case byte(RejectStreamMsgTag):
		return NewControlMsg(RejectStreamMsgTag, dataBuf)
	}
	return nil
}

func writeMsg(msgTag Type, data []byte) []byte {
	frameBuf := bytes.NewBuffer([]byte{})
	binary.Write(frameBuf, binary.BigEndian, msgTag)
	binary.Write(frameBuf, binary.BigEndian, data)
	return frameBuf.Bytes()
}

// client send RequestRawSreamFrame to apply a new stream from server
type RequestRawStreamMsg struct {
	data []byte
}

func (rdsf RequestRawStreamMsg) Type() Type {
	return RequestRawStreamMsgTag
}

func (rdsf RequestRawStreamMsg) Encode() []byte {
	return writeMsg(RequestRawStreamMsgTag, rdsf.data)
}

func (rdsf RequestRawStreamMsg) GetData() []byte {
	return rdsf.data
}

// client send RequestFrameStreamMsg to apply a new framestream from server
type RequestFrameStreamMsg struct {
	data []byte
}

func (rfsf RequestFrameStreamMsg) Type() Type {
	return RequestFrameStreamMsgTag
}

func (rfsf RequestFrameStreamMsg) Encode() []byte {
	return writeMsg(RequestFrameStreamMsgTag, rfsf.data)
}

func (rfsf RequestFrameStreamMsg) GetData() []byte {
	return rfsf.data
}

// AckStreamMsg sent from server to client after client sent RequestDawSreamFrame
type AckStreamMsg struct {
	data []byte
}

func (adsf AckStreamMsg) Type() Type {
	return AckStreamMsgTag
}

func (adsf AckStreamMsg) Encode() []byte {
	return writeMsg(AckStreamMsgTag, adsf.data)
}

func (adsf AckStreamMsg) GetData() []byte {
	return adsf.data
}

// RejectRawStreamFrame sent from server to client while occur err
type RejectStreamMsg struct {
	data []byte
}

func (rdsf RejectStreamMsg) Type() Type {
	return RejectStreamMsgTag
}

func (rdsf RejectStreamMsg) Encode() []byte {
	return writeMsg(RejectStreamMsgTag, rdsf.data)
}

func (rdsf RejectStreamMsg) GetData() []byte {
	return rdsf.data
}
