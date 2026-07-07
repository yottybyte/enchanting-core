package network

import (
	"errors"
	"io"
	"log"
	"net"

	"github.com/yottybyte/enchanting-core/internal/adapter/session"
)

func (s *Server) handleConn(conn net.Conn) {
	defer func() { _ = conn.Close() }()

	tr := s.cf(conn)
	sess := session.New(tr, s.scfg)

	for {
		id, body, err := tr.ReadFrame()
		if err != nil {
			if !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
				log.Printf("conn %s: read: %v", conn.RemoteAddr(), err)
			}
			return
		}
		if err := sess.Handle(id, body); err != nil {
			log.Printf("conn %s: handle: %v", conn.RemoteAddr(), err)
			return
		}
	}
}
