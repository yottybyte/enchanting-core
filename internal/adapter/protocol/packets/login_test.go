package packets

import (
	"bytes"
	"testing"

	_protocol2 "github.com/yottybyte/enchanting-core/internal/adapter/protocol"
)

func TestLoginStartServerbound_Decode(t *testing.T) {
	const name = "YottyByte"
	uuid := _protocol2.UUID{
		0x7d, 0x6d, 0x2e, 0x7a, 0xc6, 0x33, 0x4a, 0xb5,
		0x97, 0x20, 0x84, 0x80, 0xf4, 0x0b, 0x14, 0x5b,
	}

	w := &_protocol2.Writer{}
	w.String(name)
	w.UUID(uuid)

	r := _protocol2.NewReader(w.Bytes())

	var ls LoginStartServerbound
	ls.Decode(r)
	if err := r.Err(); err != nil {
		t.Fatalf("Decode: unexpected error: %v", err)
	}

	if ls.Name != name {
		t.Errorf("Name = %q, want %q", ls.Name, name)
	}
	if ls.UUID != uuid {
		t.Errorf("PlayerUUID = %x, want %x", ls.UUID, uuid)
	}
	if got := ls.ID(); got != 0x00 {
		t.Errorf("ID = %#x, want 0x00", got)
	}
}

func TestEncryptionRequestClientbound_Encode(t *testing.T) {
	pkt := EncryptionRequestClientbound{
		ServerID:           "",
		PublicKey:          []byte{0x30, 0x82, 0x01, 0x22},
		VerifyToken:        []byte{0xde, 0xad, 0xbe, 0xef},
		ShouldAuthenticate: true,
	}

	w := &_protocol2.Writer{}
	pkt.Encode(w)

	r := _protocol2.NewReader(w.Bytes())
	if got := r.String(); got != pkt.ServerID {
		t.Errorf("ServerID = %q, want %q", got, pkt.ServerID)
	}
	if got := r.ByteArray(); !bytes.Equal(got, pkt.PublicKey) {
		t.Errorf("PublicKey = %x, want %x", got, pkt.PublicKey)
	}
	if got := r.ByteArray(); !bytes.Equal(got, pkt.VerifyToken) {
		t.Errorf("VerifyToken = %x, want %x", got, pkt.VerifyToken)
	}
	if got := r.Bool(); got != pkt.ShouldAuthenticate {
		t.Errorf("ShouldAuthenticate = %v, want %v", got, pkt.ShouldAuthenticate)
	}
	if err := r.Err(); err != nil {
		t.Fatalf("reader err: %v", err)
	}
	if pkt.ID() != 0x01 {
		t.Errorf("ID = %#x, want 0x01", pkt.ID())
	}
}

func TestLoginSuccessClientbound_Encode(t *testing.T) {
	pkt := LoginSuccessClientbound{
		UUID:       _protocol2.UUID{0x7d, 0x6d, 0x2e, 0x7a, 0xc6, 0x33, 0x4a, 0xb5, 0x97, 0x2, 0xf4, 0x0b, 0x14, 0x5b},
		Username:   "YottyByte",
		Properties: []LoginProperty{{Name: "textures", Value: "v", Signature: "s"}},
	}
	w := &_protocol2.Writer{}
	pkt.Encode(w)

	r := _protocol2.NewReader(w.Bytes())
	if r.UUID() != pkt.UUID {
		t.Error("uuid mismatch")
	}
	if r.String() != pkt.Username {
		t.Error("username mismatch")
	}
	if n := r.VarInt(); n != 1 {
		t.Fatalf("properties count = %d, want 1", n)
	}
	if r.String() != "textures" || r.String() != "v" {
		t.Error("property name/value mismatch")
	}
	if !r.Bool() || r.String() != "s" {
		t.Error("signature mismatch")
	}
	if err := r.Err(); err != nil {
		t.Fatalf("reader err: %v", err)
	}
	if pkt.ID() != 0x02 {
		t.Errorf("ID = %#x, want 0x02", pkt.ID())
	}
}
