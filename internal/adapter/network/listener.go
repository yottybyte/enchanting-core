package network

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/yottybyte/enchanting-core/internal/adapter/auth"
	"github.com/yottybyte/enchanting-core/internal/adapter/session"
	"github.com/yottybyte/enchanting-core/internal/adapter/status"
	"github.com/yottybyte/enchanting-core/internal/config"
	"github.com/yottybyte/enchanting-core/internal/domain"
)

type Server struct {
	address string
	ln      net.Listener
	wg      sync.WaitGroup

	cf   ConnectionTransportFactory
	cfg  *config.Server
	scfg *session.Config
}

func NewServer(cfg *config.Config, cf ConnectionTransportFactory) *Server {
	ss := domain.ServerStatus{
		MaxPlayers:    cfg.Server.MaxPlayers,
		OnlinePlayers: 0,
		Description:   cfg.Server.MOTD,
	}
	statusJSON, err := status.NewStatus(ss).Build()
	if err != nil {
		return nil
	}

	return &Server{
		address: fmt.Sprintf(":%d", cfg.Server.Port),
		cf:      cf,
		cfg:     &cfg.Server,
		scfg: &session.Config{
			StatusJSON: statusJSON,
			Auth:       auth.NewClient(),
		}}
}

func (s *Server) Listen() error {
	ln, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}
	s.ln = ln
	return nil
}

func (s *Server) Serve(ctx context.Context) error {
	defer func() { _ = s.ln.Close() }()

	stop := context.AfterFunc(ctx, func() { _ = s.ln.Close() })
	defer stop()

	for {
		conn, err := s.ln.Accept()
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				log.Println("accept:", err)
			}
			break
		}
		s.wg.Go(func() { s.handleConn(conn) })
	}

	s.wg.Wait()
	return nil
}

func (s *Server) Run(ctx context.Context) error {
	if s.cfg.OnlineMode {
		if err := s.generateEncryptionKeys(); err != nil {
			return err
		}
	}

	if err := s.Listen(); err != nil {
		return err
	}
	return s.Serve(ctx)
}

func (s *Server) Addr() net.Addr {
	return s.ln.Addr()
}

func (s *Server) generateEncryptionKeys() error {
	var err error
	s.scfg.PrivateKey, err = rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return err
	}

	s.scfg.PublicKeyDER, err = x509.MarshalPKIXPublicKey(&s.scfg.PrivateKey.PublicKey)
	if err != nil {
		return err
	}

	return nil
}
