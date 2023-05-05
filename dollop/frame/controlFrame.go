package frame

// client send RequestRawSreamFrame to apply a new stream from server
type RequestRawStreamFrame struct {
	data []byte
}

func (rdsf RequestRawStreamFrame) Type() Type {
	return RequestRawStreamFrameTag
}

func (rdsf RequestRawStreamFrame) Encode() []byte {
	return WriteFrame(rdsf.data, RequestRawStreamFrameTag)
}

func (rdsf RequestRawStreamFrame) GetData() []byte {
	return rdsf.data
}

func NewRequestRawStreamFrame() RequestRawStreamFrame {
	return RequestRawStreamFrame{}
}

func ParseRequestRawStreamFrame(data []byte) (RequestRawStreamFrame, error) {
	return RequestRawStreamFrame{data: data}, nil
}

// client send RequestFrameStreamFrame to apply a new framestream from server
type RequestFrameStreamFrame struct {
	data []byte
}

func (rfsf RequestFrameStreamFrame) Type() Type {
	return RequestFrameStreamFrameTag
}

func (rfsf RequestFrameStreamFrame) Encode() []byte {
	return WriteFrame(rfsf.data, RequestFrameStreamFrameTag)
}

func (rfsf RequestFrameStreamFrame) GetData() []byte {
	return rfsf.data
}

func NewRequestFrameStreamFrame() RequestFrameStreamFrame {
	return RequestFrameStreamFrame{}
}

func ParseRequestFrameStreamFrame(data []byte) (RequestFrameStreamFrame, error) {
	return RequestFrameStreamFrame{data: data}, nil
}

// AckStreamFrame sent from server to client after client sent RequestDawSreamFrame
type AckStreamFrame struct {
	data []byte
}

func (adsf AckStreamFrame) Type() Type {
	return AckStreamFrameTag
}

func (adsf AckStreamFrame) Encode() []byte {
	return WriteFrame(adsf.data, AckStreamFrameTag)
}

func (adsf AckStreamFrame) GetData() []byte {
	return adsf.data
}

func NewAckStreamFrame() AckStreamFrame {
	return AckStreamFrame{}
}

func ParseAckStreamFrame(data []byte) (AckStreamFrame, error) {
	return AckStreamFrame{data: data}, nil
}

// RejectRawStreamFrame sent from server to client while occur err
type RejectStreamFrame struct {
	data []byte
}

func (rdsf RejectStreamFrame) Type() Type {
	return RejectStreamFrameTag
}

func (rdsf RejectStreamFrame) Encode() []byte {
	return WriteFrame(rdsf.data, RejectStreamFrameTag)
}

func (rdsf RejectStreamFrame) GetData() []byte {
	return rdsf.data
}

func NewRejectStreamFrame() RejectStreamFrame {
	return RejectStreamFrame{}
}

func ParseRejectStreamFrame(data []byte) (RejectStreamFrame, error) {
	return RejectStreamFrame{data: data}, nil
}
