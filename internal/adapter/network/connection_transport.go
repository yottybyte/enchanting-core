package network

import (
	"net"

	"github.com/yottybyte/enchanting-core/internal/adapter/protocol"
)

type ConnectionTransportFactory func(conn net.Conn) ConnectionTransport

type ConnectionTransport interface {
	ReadFrame() (int32, *protocol.Reader, error)
	Write([]byte) (int, error)
	EnableEncryption([]byte) error
}
