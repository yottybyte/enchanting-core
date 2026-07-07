package session

import (
	"github.com/yottybyte/enchanting-core/internal/adapter/protocol"
)

type PacketHandler interface {
	ID() int32
	Handle(Ctx, *protocol.Reader) error
}
