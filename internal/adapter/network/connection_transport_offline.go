package network

import (
	"bufio"
	"io"
	"net"

	_protocol2 "github.com/yottybyte/enchanting-core/internal/adapter/protocol"
)

type ConnOfflineTransport struct {
	conn net.Conn
	br   *bufio.Reader
	w    io.Writer
}

func NewConnOfflineTransport() ConnectionTransportFactory {
	return func(conn net.Conn) ConnectionTransport {
		return &ConnOfflineTransport{conn: conn, br: bufio.NewReader(conn), w: conn}
	}
}

func (t *ConnOfflineTransport) ReadFrame() (int32, *_protocol2.Reader, error) {
	return _protocol2.ReadFrame(t.br)
}

func (t *ConnOfflineTransport) Write(p []byte) (int, error) {
	return t.w.Write(p)
}

func (t *ConnOfflineTransport) EnableEncryption(_ []byte) error {
	return nil
}
