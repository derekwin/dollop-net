package dollop

import "github.com/quic-go/quic-go"

type RequestI interface {
	GetConn() ConnectionI
	GetStream() quic.Stream
	GetData() []byte
}

// Request bind stream with data

type Request struct {
	conn   ConnectionI
	stream quic.Stream
	data   []byte
}

func (r Request) GetConn() ConnectionI {
	return r.conn
}

func (r Request) GetStream() quic.Stream {
	return r.stream
}

func (r Request) GetData() []byte {
	return r.data
}
