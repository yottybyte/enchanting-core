package session

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/subtle"
	"fmt"
	"io"
	"log"

	_protocol2 "github.com/yottybyte/enchanting-core/internal/adapter/protocol"
	packets2 "github.com/yottybyte/enchanting-core/internal/adapter/protocol/packets"
	"github.com/yottybyte/enchanting-core/internal/domain"
)

const serverID = "dev-enchanting"

type Authenticator interface {
	HasJoined(ctx context.Context, serverID, username string, sharedSecret, publicKey []byte) (*domain.Profile, error)
}

type Conn interface {
	io.Writer
	EnableEncryption(secret []byte) error
}

type Session struct {
	conn  Conn
	state _protocol2.State

	username     string
	verifyToken  []byte
	sharedSecret []byte

	config *Config

	profile *domain.Profile
}

type Config struct {
	StatusJSON   string
	PrivateKey   *rsa.PrivateKey
	PublicKeyDER []byte

	Auth Authenticator
}

func New(conn Conn, config *Config) *Session {
	return &Session{conn: conn, state: _protocol2.StateHandshake, config: config}
}

func (s *Session) send(pkt _protocol2.Clientbound) error {
	return _protocol2.WritePacket(s.conn, pkt)
}

func (s *Session) Handle(id int32, body *_protocol2.Reader) error {
	switch s.state {
	case _protocol2.StateHandshake:
		return s.handleHandshake(id, body)
	case _protocol2.StateStatus:
		return s.handleStatus(id, body)
	case _protocol2.StateLogin:
		return s.handleLogin(id, body)
	case _protocol2.StateConfiguration:
		return s.handleConfiguration(id, body)
	default:
		return fmt.Errorf("session: no handler for state %d", s.state)
	}
}

func (s *Session) handleConfiguration(id int32, body *_protocol2.Reader) error {
	log.Printf("session: configuration packet id %d not implemented", id)
	return nil
}

func (s *Session) handleLogin(id int32, body *_protocol2.Reader) error {
	switch id {
	case 0x00:
		var ls packets2.LoginStartServerbound
		ls.Decode(body)
		if body.Err() != nil {
			return body.Err()
		}

		s.username = ls.Name

		token := make([]byte, 4)
		if _, err := rand.Read(token); err != nil {
			return err
		}
		s.verifyToken = token

		return s.send(&packets2.EncryptionRequestClientbound{
			ServerID:           serverID,
			PublicKey:          s.config.PublicKeyDER,
			VerifyToken:        s.verifyToken,
			ShouldAuthenticate: true,
		})
	case 0x01:
		var er packets2.EncryptionResponseServerbound
		er.Decode(body)
		if body.Err() != nil {
			return body.Err()
		}

		//nolint:staticcheck // The MC protocol requires PKCS#1 v1.5; OAEP is incompatible with
		secret, err := rsa.DecryptPKCS1v15(nil, s.config.PrivateKey, er.SharedSecret)
		if err != nil {
			return err
		}

		if len(secret) != 16 {
			return fmt.Errorf("session: invalid secret length %d", len(secret))
		}

		//nolint:staticcheck //The MC protocol requires PKCS#1 v1.5; OAEP is incompatible with
		verifyToken, err := rsa.DecryptPKCS1v15(nil, s.config.PrivateKey, er.VerifyToken)
		if err != nil {
			return err
		}

		if subtle.ConstantTimeCompare(verifyToken, s.verifyToken) != 1 {
			return fmt.Errorf("session: verify token mismatch")
		}

		s.sharedSecret = secret
		if err := s.conn.EnableEncryption(secret); err != nil {
			return err
		}

		profile, err := s.config.Auth.HasJoined(context.Background(), serverID, s.username, s.sharedSecret, s.config.PublicKeyDER)
		if err != nil {
			return err
		}
		s.profile = profile
		log.Printf("authenticated: name=%s uuid=%s", profile.Name, profile.ID)

		uuid, err := _protocol2.ParseUUID(profile.ID)
		if err != nil {
			return err
		}

		prop := make([]packets2.LoginProperty, len(profile.Properties))
		for i, p := range profile.Properties {
			prop[i] = packets2.LoginProperty{Name: p.Name, Value: p.Value, Signature: p.Signature}
		}

		var sid _protocol2.UUID
		if _, err := rand.Read(sid[:]); err != nil {
			return err
		}
		sid[6] = (sid[6] & 0x0f) | 0x40
		sid[8] = (sid[8] & 0x3f) | 0x80

		return s.send(&packets2.LoginSuccessClientbound{
			UUID:       uuid,
			Username:   profile.Name,
			Properties: nil,
			SessionID:  sid,
		})
	case 0x03:
		s.state = _protocol2.StateConfiguration
		return nil
	default:
		return fmt.Errorf("session: unexpected login packet id %d", id)
	}

}

func (s *Session) handleHandshake(id int32, body *_protocol2.Reader) error {
	if id != 0x00 {
		return fmt.Errorf("session: unexpected handshake packet id %d", id)
	}

	var h packets2.HandshakeServerbound
	h.Decode(body)
	if err := body.Err(); err != nil {
		return err
	}

	next, err := _protocol2.StateFromNext(h.NextState)
	if err != nil {
		return err
	}

	s.state = next
	return nil
}

func (s *Session) handleStatus(id int32, body *_protocol2.Reader) error {
	switch id {
	case 0x00:
		return s.send(&packets2.StatusResponseClientbound{
			JSONResponse: s.config.StatusJSON,
		})
	case 0x01:
		var pkt packets2.PingRequestServerbound
		pkt.Decode(body)
		if body.Err() != nil {
			return body.Err()
		}
		return s.send(&packets2.PongResponseClientbound{Timestamp: pkt.Timestamp})
	default:
		return fmt.Errorf("session: unexpected packet id %d", id)
	}
}
