package dollop

import (
	"github.com/quic-go/quic-go"
)

type RawStreamI interface {
	quic.Stream
}

// RawStream is equal to quic.Stream.
type RawStream struct {
}

// NewFrameStream creates a new FrameStream.
func NewRawStream(s quic.Stream) RawStreamI {
	return s
}
