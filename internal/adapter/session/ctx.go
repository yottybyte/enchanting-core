package session

import (
	"crypto/rsa"

	"github.com/yottybyte/enchanting-core/internal/adapter/protocol"
	"github.com/yottybyte/enchanting-core/internal/domain"
)

type Ctx interface {
	Enqueue(protocol.Clientbound)
	EnableEncryption(secret []byte) error
	SetState(protocol.State)

	Login() *domain.LoginState

	PrivateKey() *rsa.PrivateKey
	PublicKeyDer() []byte
	StatusJson() string
	ServerID() string
	Auth() Authenticator
}
