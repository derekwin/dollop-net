package dollop

import (
	"github.com/quic-go/quic-go"
)

// RawStream is equal to quic.Stream.
type RawStreamI interface {
	StreamID() StreamID
	BindRawRouter(r RawRouterI)
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Close() error
}

type RawStream struct {
	stream  quic.Stream
	routers []RawRouterI
}

// NewFrameStream creates a new FrameStream.
func NewRawStream(s quic.Stream) *RawStream {
	return &RawStream{stream: s}
}

func (rs *RawStream) BindRawRouter(r RawRouterI) {
	rs.routers = append(rs.routers, r)
}

func (rs *RawStream) Read(p []byte) (n int, err error) {
	return rs.stream.Read(p)
}

func (rs *RawStream) Write(p []byte) (n int, err error) {
	return rs.stream.Write(p)
}

func (rs *RawStream) Close() error {
	return rs.stream.Close()
}

func (rs *RawStream) StreamID() StreamID {
	return StreamID(rs.stream.StreamID())
}
