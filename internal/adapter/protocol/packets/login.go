package packets

import (
	_protocol2 "github.com/yottybyte/enchanting-core/internal/adapter/protocol"
)

var _ _protocol2.Serverbound = (*LoginStartServerbound)(nil)
var _ _protocol2.Serverbound = (*EncryptionResponseServerbound)(nil)

var _ _protocol2.Clientbound = (*EncryptionRequestClientbound)(nil)
var _ _protocol2.Clientbound = (*DisconnectClientbound)(nil)
var _ _protocol2.Clientbound = (*LoginSuccessClientbound)(nil)

type LoginStartServerbound struct {
	Name string
	UUID _protocol2.UUID
}

func (l *LoginStartServerbound) ID() int32 {
	return 0x0
}

func (l *LoginStartServerbound) Decode(reader *_protocol2.Reader) {
	l.Name = reader.String()
	l.UUID = reader.UUID()
}

type EncryptionResponseServerbound struct {
	SharedSecret []byte
	VerifyToken  []byte
}

func (e *EncryptionResponseServerbound) ID() int32 {
	return 0x1
}

func (e *EncryptionResponseServerbound) Decode(r *_protocol2.Reader) {
	e.SharedSecret = r.ByteArray()
	e.VerifyToken = r.ByteArray()
}

type DisconnectClientbound struct {
	Reason string
}

func (d *DisconnectClientbound) ID() int32 {
	return 0x0
}

func (d *DisconnectClientbound) Encode(w *_protocol2.Writer) {
	w.String(d.Reason)
}

type EncryptionRequestClientbound struct {
	ServerID           string
	PublicKey          []byte
	VerifyToken        []byte
	ShouldAuthenticate bool
}

func (e *EncryptionRequestClientbound) ID() int32 {
	return 0x1
}

func (e *EncryptionRequestClientbound) Encode(w *_protocol2.Writer) {
	w.String(e.ServerID)
	w.ByteArray(e.PublicKey)
	w.ByteArray(e.VerifyToken)
	w.Bool(e.ShouldAuthenticate)
}

type LoginProperty struct {
	Name      string
	Value     string
	Signature string
}

type LoginSuccessClientbound struct {
	UUID       _protocol2.UUID
	Username   string
	Properties []LoginProperty
	SessionID  _protocol2.UUID
}

func (l *LoginSuccessClientbound) ID() int32 {
	return 0x02
}

func (l *LoginSuccessClientbound) Encode(w *_protocol2.Writer) {
	w.UUID(l.UUID)
	w.String(l.Username)
	w.VarInt(int32(len(l.Properties)))
	for _, prop := range l.Properties {
		w.String(prop.Name)
		w.String(prop.Value)
		w.Bool(prop.Signature != "")
		if prop.Signature != "" {
			w.String(prop.Signature)
		}
	}
	w.UUID(l.SessionID)
}
