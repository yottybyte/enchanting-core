package packets

import (
	_protocol2 "github.com/yottybyte/enchanting-core/internal/adapter/protocol"
)

var _ _protocol2.Serverbound = (*HandshakeServerbound)(nil)

type HandshakeServerbound struct {
	ProtocolVersion int32
	ServerAddress   string
	ServerPort      uint16
	NextState       int32
}

func (h *HandshakeServerbound) ID() int32 {
	return 0x00
}

func (h *HandshakeServerbound) Decode(r *_protocol2.Reader) {
	h.ProtocolVersion = r.VarInt()
	h.ServerAddress = r.String()
	h.ServerPort = r.U16()
	h.NextState = r.VarInt()
}
