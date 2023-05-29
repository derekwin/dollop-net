package dollop

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// default uint8,
// can set any type for this field
type MsgType interface{}

type MsgI interface {
	Type() MsgType
	// Encode the frame into []byte.
	Encode() []byte

	GetData() []byte
}

// TODO，实现抽象的MsgI层，绑定到子流上，路由直接处理msg；msgProtocol是消息到路由的映射
type MsgProtocolI interface {
	Name() string
	Version() string
	PaserMsg(f *Frame) MsgI
	AddM2R(tag MsgType, router FrameRouterI) error
	GetRouter(tag MsgType) (FrameRouterI, error)
}

// common function: build msg
func BuildMsg(msgTag MsgType, data []byte) []byte {
	frameBuf := bytes.NewBuffer([]byte{})
	binary.Write(frameBuf, binary.BigEndian, msgTag)
	binary.Write(frameBuf, binary.BigEndian, data)
	return frameBuf.Bytes()
}

// default Base Msg protocol
type BaseMsgType uint8

// BaseType's byte len : uint8 -> 1
const BaseMsgTypeLen int = 1

const (
	BaseMsgTag BaseMsgType = 0x01
)

// base Msg
type BaseMsg struct {
	data []byte
}

func (bm BaseMsg) Type() MsgType {
	return BaseMsgTag // 以interface的形式传出对应Tag
}

func (bm BaseMsg) Encode() []byte {
	return BuildMsg(BaseMsgTag, bm.data)
}

func (bm BaseMsg) GetData() []byte {
	return bm.data
}

func NewBaseMsg(data []byte) *BaseMsg {
	return &BaseMsg{data: data}
}

// base Msg protocol
type BaseMsgProtocol struct {
	name    string
	version string
	M2R     map[BaseMsgType]FrameRouterI
}

func (bmp BaseMsgProtocol) Name() string {
	return bmp.name
}

func (bmp BaseMsgProtocol) Version() string {
	return bmp.version
}

func (bmp BaseMsgProtocol) PaserMsg(f *Frame) MsgI {
	msgType := f.data[0]
	dataBuf := f.data[BaseMsgTypeLen:]

	switch msgType {
	case byte(BaseMsgTag):
		return NewBaseMsg(dataBuf)
	}
	return nil
}

func (bmp *BaseMsgProtocol) AddM2R(tag MsgType, router FrameRouterI) error {
	t := tag.(BaseMsgType)
	bmp.M2R[t] = router
	return nil
}

func (bmp BaseMsgProtocol) GetRouter(tag MsgType) (FrameRouterI, error) {
	t := tag.(BaseMsgType)
	r := bmp.M2R[t]
	if r != nil {
		return bmp.M2R[t], nil
	}
	return nil, fmt.Errorf("msg has not a valid router")
}

var defaultMsgProtocol = &BaseMsgProtocol{
	name:    "defaultMsgProtocol",
	version: "v0",
	M2R: map[BaseMsgType]FrameRouterI{
		BaseMsgTag: BaseFrameRouter{},
	},
}
