package packets

import (
	_protocol2 "github.com/yottybyte/enchanting-core/internal/adapter/protocol"
)

var _ _protocol2.Serverbound = (*StatusRequestServerbound)(nil)
var _ _protocol2.Serverbound = (*PingRequestServerbound)(nil)
var _ _protocol2.Clientbound = (*StatusResponseClientbound)(nil)
var _ _protocol2.Clientbound = (*PongResponseClientbound)(nil)

type StatusRequestServerbound struct{}

func (s *StatusRequestServerbound) ID() int32 {
	return 0x0
}

func (s *StatusRequestServerbound) Decode(_ *_protocol2.Reader) {}

type PingRequestServerbound struct {
	Timestamp int64
}

func (p *PingRequestServerbound) ID() int32 {
	return 0x1
}

func (p *PingRequestServerbound) Decode(r *_protocol2.Reader) {
	p.Timestamp = r.I64()
}

type StatusResponseClientbound struct {
	JSONResponse string
}

func (s *StatusResponseClientbound) ID() int32 {
	return 0x0
}

func (s *StatusResponseClientbound) Encode(w *_protocol2.Writer) {
	w.String(s.JSONResponse)
}

type PongResponseClientbound struct {
	Timestamp int64
}

func (p *PongResponseClientbound) ID() int32 {
	return 0x1
}

func (p *PongResponseClientbound) Encode(w *_protocol2.Writer) {
	w.I64(p.Timestamp)
}
