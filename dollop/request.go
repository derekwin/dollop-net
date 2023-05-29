package dollop

type RequestI interface {
	GetConn() (ConnectionI, error)
	GetData() ([]byte, error)
}

type RawRequestI interface {
	RequestI
	GetStream() (RawStreamI, error)
}

// Request bind stream with data

type RawRequest struct {
	conn   ConnectionI
	stream RawStreamI
	data   []byte
}

func (r RawRequest) GetConn() (ConnectionI, error) {
	return r.conn, nil
}

func (r RawRequest) GetStream() (RawStreamI, error) {
	return r.stream, nil
}

func (r RawRequest) GetData() ([]byte, error) {
	return r.data, nil
}

type FrameRequestI interface {
	RequestI
	GetStream() (FrameStreamI, error)
}

// Request bind stream with data

type FrameRequest struct {
	conn   ConnectionI
	stream FrameStreamI
	msg    MsgI
}

func (r FrameRequest) GetConn() (ConnectionI, error) {
	return r.conn, nil
}

func (r FrameRequest) GetStream() (FrameStreamI, error) {
	return r.stream, nil
}

func (r FrameRequest) GetMsg() (MsgI, error) {
	return r.msg, nil
}

func (r FrameRequest) GetData() ([]byte, error) {
	return r.msg.GetData(), nil
}
