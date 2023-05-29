package dollop

import (
	"context"
	"crypto/tls"

	"github.com/quic-go/quic-go"
)

// Client
type Client struct {
	// server name
	Name string

	QuicConfig *quic.Config
	TlsConfig  *tls.Config
	// logger     *slog.Logger
	conn *ClientConnection
}

func NewClient(name string, tlsConfig *tls.Config, qConf *quic.Config) *Client {
	return &Client{Name: name, TlsConfig: tlsConfig, QuicConfig: qConf}
}

func (c *Client) Connect(addr string) {
	conn, err := quic.DialAddr(addr, c.TlsConfig, c.QuicConfig)
	if err != nil {
		panic(err)
	}

	c.conn = NewClientConnection(context.Background(), conn)

	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		panic(err)
	}

	controlStream := NewFrameStream(stream)
	controlStream.BindMsgProtocol(controlMsgProtocol)

	c.conn.setControlStream(controlStream)
}

func (c *Client) NewRawStream() (RawStreamI, StreamID, error) {
	return c.conn.OpenNewRawStream()
}

func (c *Client) GetRawStream(id StreamID) (RawStreamI, error) {
	return c.conn.GetRawStream(id)
}

func (c *Client) NewFrameStream() (FrameStreamI, StreamID, error) {
	return c.conn.OpenNewFrameStream()
}

func (c *Client) GetFrameStream(id StreamID) (FrameStreamI, error) {
	return c.conn.GetFrameStream(id)
}
