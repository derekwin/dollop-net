package dollop

import "github.com/quic-go/quic-go"

type RequestI interface {
	GetConn() ConnectionI
	GetData() []byte
}

type RawRequestI interface {
	RequestI
	GetStream() quic.Stream
}

// Request bind stream with data

type RawRequest struct {
	conn   ConnectionI
	stream quic.Stream
	data   []byte
}

func (r RawRequest) GetConn() ConnectionI {
	return r.conn
}

func (r RawRequest) GetStream() quic.Stream {
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
