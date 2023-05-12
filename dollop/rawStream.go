package dollop

import (
	"github.com/quic-go/quic-go"
)

// RawStream is equal to quic.Stream.
type RawStreamI interface {
	quic.Stream
}

type RawStream struct {
}

// NewFrameStream creates a new FrameStream.
func NewRawStream(s RawStreamI) RawStreamI {
	return s
}
