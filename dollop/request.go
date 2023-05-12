package dollop

type RequestI interface {
	GetConn() ConnectionI
	GetData() []byte
}

type RawRequestI interface {
	RequestI
	GetStream() RawStreamI
}

// Request bind stream with data

type RawRequest struct {
	conn   ConnectionI
	stream RawStreamI
	data   []byte
}

func (r RawRequest) GetConn() ConnectionI {
	return r.conn
}

func (r RawRequest) GetStream() RawStreamI {
	return r.stream
}

func (r RawRequest) GetData() []byte {
	return r.data
}

type FrameRequestI interface {
	RequestI
	GetStream() FrameStreamI
}

// Request bind stream with data

type FrameRequest struct {
	conn   ConnectionI
	stream FrameStreamI
	data   []byte
}

func (r FrameRequest) GetConn() ConnectionI {
	return r.conn
}

func (r FrameRequest) GetStream() FrameStreamI {
	return r.stream
}

func (r FrameRequest) GetData() []byte {
	return r.data
}
