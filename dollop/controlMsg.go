package dollop

import "fmt"

type ControlMsgType uint8

const ControlMsgTypeLen int = 1 // Type's byte len : uint8 -> 1

const (
	// control Frame Tag
	RequestRawStreamMsgTag   ControlMsgType = 0x01
	RequestFrameStreamMsgTag ControlMsgType = 0x02
	AckStreamMsgTag          ControlMsgType = 0x03
	RejectStreamMsgTag       ControlMsgType = 0x04
)

// client send RequestRawSreamFrame to apply a new stream from server
type RequestRawStreamMsg struct {
	data []byte
}

func (rdsf RequestRawStreamMsg) Type() MsgType {
	return RequestRawStreamMsgTag
}

func (rdsf RequestRawStreamMsg) Encode() []byte {
	return BuildMsg(RequestRawStreamMsgTag, rdsf.data)
}

func (rdsf RequestRawStreamMsg) GetData() []byte {
	return rdsf.data
}

func NewRequestRawStreamMsg(data []byte) *RequestRawStreamMsg {
	return &RequestRawStreamMsg{data: data}
}

// client send RequestFrameStreamMsg to apply a new framestream from server
type RequestFrameStreamMsg struct {
	data []byte
}

func (rfsf RequestFrameStreamMsg) Type() MsgType {
	return RequestFrameStreamMsgTag
}

func (rfsf RequestFrameStreamMsg) Encode() []byte {
	return BuildMsg(RequestFrameStreamMsgTag, rfsf.data)
}

func (rfsf RequestFrameStreamMsg) GetData() []byte {
	return rfsf.data
}

func NewRequestFrameStreamMsg(data []byte) *RequestFrameStreamMsg {
	return &RequestFrameStreamMsg{data: data}
}

// AckStreamMsg sent from server to client after client sent RequestDawSreamFrame
type AckStreamMsg struct {
	data []byte
}

func (adsf AckStreamMsg) Type() MsgType {
	return AckStreamMsgTag
}

func (adsf AckStreamMsg) Encode() []byte {
	return BuildMsg(AckStreamMsgTag, adsf.data)
}

func (adsf AckStreamMsg) GetData() []byte {
	return adsf.data
}

func NewAckStreamMsg(data []byte) *AckStreamMsg {
	return &AckStreamMsg{data: data}
}

// RejectRawStreamFrame sent from server to client while occur err
type RejectStreamMsg struct {
	data []byte
}

func (rdsf RejectStreamMsg) Type() MsgType {
	return RejectStreamMsgTag
}

func (rdsf RejectStreamMsg) Encode() []byte {
	return BuildMsg(RejectStreamMsgTag, rdsf.data)
}

func (rdsf RejectStreamMsg) GetData() []byte {
	return rdsf.data
}

func NewRejectStreamMsg(data []byte) *RejectStreamMsg {
	return &RejectStreamMsg{data: data}
}

// base Msg protocol
type ControlMsgProtocol struct {
	name    string
	version string
	M2R     map[ControlMsgType]FrameRouterI
}

func (cmp ControlMsgProtocol) Name() string {
	return cmp.name
}

func (cmp ControlMsgProtocol) Version() string {
	return cmp.version
}

func (cmp ControlMsgProtocol) PaserMsg(f *Frame) MsgI {
	msgType := f.data[0]
	dataBuf := f.data[ControlMsgTypeLen:]

	switch msgType {
	case byte(RequestRawStreamMsgTag):
		return NewRequestRawStreamMsg(dataBuf)
	case byte(RequestFrameStreamMsgTag):
		return NewRequestFrameStreamMsg(dataBuf)
	case byte(AckStreamMsgTag):
		return NewAckStreamMsg(dataBuf)
	case byte(RejectStreamMsgTag):
		return NewRejectStreamMsg(dataBuf)
	}
	return nil
}

func (bmp *ControlMsgProtocol) AddM2R(tag MsgType, router FrameRouterI) error {
	t := tag.(ControlMsgType)
	bmp.M2R[t] = router
	return nil
}

func (bmp ControlMsgProtocol) GetRouter(tag MsgType) (FrameRouterI, error) {
	t := tag.(ControlMsgType)
	r := bmp.M2R[t]
	if r != nil {
		return bmp.M2R[t], nil
	}
	return nil, fmt.Errorf("msg has not a valid router")
}

func ParseControlMsg(frame *Frame) MsgI {

	msgType := frame.data[0]
	dataBuf := frame.data[ControlMsgTypeLen:]

	switch msgType {
	case byte(RequestRawStreamMsgTag):
		return NewRequestRawStreamMsg(dataBuf)
	case byte(RequestFrameStreamMsgTag):
		return NewRequestFrameStreamMsg(dataBuf)
	case byte(AckStreamMsgTag):
		return NewAckStreamMsg(dataBuf)
	case byte(RejectStreamMsgTag):
		return NewRejectStreamMsg(dataBuf)
	}
	return nil
}

var controlMsgProtocol = &ControlMsgProtocol{
	name:    "defaultControl",
	version: "v0",
	M2R: map[ControlMsgType]FrameRouterI{
		RequestRawStreamMsgTag:   &RequestRawStreamRouter{},
		RequestFrameStreamMsgTag: &RequestFrameStreamRouter{},
	},
}
