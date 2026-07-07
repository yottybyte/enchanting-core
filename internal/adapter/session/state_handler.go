package session

import (
	_protocol2 "github.com/yottybyte/enchanting-core/internal/adapter/protocol"
)

type StateHandler interface {
	State() _protocol2.State
	Handle(Ctx, *_protocol2.Reader) error
}
