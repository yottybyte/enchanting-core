package network

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"
	"net"

	_protocol2 "github.com/yottybyte/enchanting-core/internal/adapter/protocol"
)

type ConnOnlineTransport struct {
	conn net.Conn
	br   *bufio.Reader
	w    io.Writer
}

func NewConnOnlineTransport() ConnectionTransportFactory {
	return func(conn net.Conn) ConnectionTransport {
		return &ConnOnlineTransport{conn: conn, br: bufio.NewReader(conn), w: conn}
	}
}

func (t *ConnOnlineTransport) ReadFrame() (int32, *_protocol2.Reader, error) {
	return _protocol2.ReadFrame(t.br)
}

func (t *ConnOnlineTransport) Write(p []byte) (int, error) {
	return t.w.Write(p)
}

func (t *ConnOnlineTransport) EnableEncryption(secret []byte) error {
	if n := t.br.Buffered(); n != 0 {
		return fmt.Errorf("network: %d buffered bytes before encryption", n)
	}
	block, err := aes.NewCipher(secret)
	if err != nil {
		return err
	}
	t.br = bufio.NewReader(cipher.StreamReader{S: _protocol2.NewCFB8Decrypter(block, secret), R: t.conn})
	t.w = cipher.StreamWriter{S: _protocol2.NewCFB8Encrypter(block, secret), W: t.conn}
	return nil
}
