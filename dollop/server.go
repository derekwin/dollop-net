package dollop

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"crypto/tls"

	"github.com/quic-go/quic-go"
	// "golang.org/x/exp/slog"
)

// DefalutQuicConfig be used when `quicConfig` is nil.
var DefalutQuicConfig = &quic.Config{
	Versions:                       []quic.VersionNumber{quic.VersionDraft29, quic.Version1, quic.Version2},
	MaxIdleTimeout:                 time.Second * 5,
	KeepAlivePeriod:                time.Second * 2,
	MaxIncomingStreams:             1000,
	MaxIncomingUniStreams:          1000,
	HandshakeIdleTimeout:           time.Second * 3,
	InitialStreamReceiveWindow:     1024 * 1024 * 2,
	InitialConnectionReceiveWindow: 1024 * 1024 * 2,
	// DisablePathMTUDiscovery:        true,
}

type WithConfig func(o *Server)

func WithQuicConfig(qc *quic.Config) WithConfig {
	return func(o *Server) {
		o.QuicConfig = qc
	}
}

func WithTlsConfig(tc *tls.Config) WithConfig {
	return func(o *Server) {
		o.TlsConfig = tc
	}
}

type FrameHandler func(c *context.Context) error
type ConnectionHandler func(conn quic.Connection)

// Server
type Server struct {
	// server name
	Name   string
	closed bool

	QuicConfig *quic.Config
	TlsConfig  *tls.Config

	// startHandlers           []FrameHandler
	// beforeHandlers          []FrameHandler
	// afterHandlers           []FrameHandler
	// connectionCloseHandlers []ConnectionHandler

	Router   Router
	Listener quic.Listener
	// logger     *slog.Logger

	mutex sync.Mutex
}

func NewServer(name string, opts ...WithConfig) (*Server, error) {
	s := &Server{
		Name:       name,
		closed:     false,
		TlsConfig:  nil,
		QuicConfig: DefalutQuicConfig,
	}

	for _, configFunc := range opts {
		configFunc(s)
	}

	// The tls.Config must not be nil and must contain a certificate configuration.
	if s.TlsConfig == nil {
		return &Server{}, errors.New("the tls.Config must not be nil and must contain a certificate configuration")
	}

	return s, nil
}

func (s *Server) Serve(ctx context.Context, addr string) error {

	s.mutex.Lock()
	closed := s.closed
	s.mutex.Unlock()
	if closed {
		return errors.New("err server closed")
	}

	listener, err := quic.ListenAddr(addr, s.TlsConfig, s.QuicConfig)
	if err != nil {
		fmt.Println("failed to listen on quic", err)
		return err
	}
	s.Listener = listener

	fmt.Println(s.Name, "is up and running", "pid", os.Getpid())

	for {
		qconn, err := s.Listener.Accept(ctx)
		if err != nil {
			fmt.Println(err)
			continue
		}

		conn := NewServerConnection(ctx, qconn)

		go func(conn *ServerConnection) {

			defer conn.Close()

			select {
			case <-ctx.Done():
				return
			case <-conn.Serve(ctx):
			}
		}(conn)

	}
}

func (s *Server) Stop() error {
	s.Listener.Close()
	return nil
}
