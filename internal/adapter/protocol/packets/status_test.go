package packets

import (
	"testing"

	"github.com/yottybyte/enchanting-core/internal/adapter/protocol"
)

func TestPingRequestServerboundDecode(t *testing.T) {
	const want = int64(0x0123456789ABCDEF)

	w := &protocol.Writer{}
	w.I64(want)

	var got PingRequestServerbound
	r := protocol.NewReader(w.Bytes())
	got.Decode(r)

	if err := r.Err(); err != nil {
		t.Fatalf("Decode: unexpected error: %v", err)
	}
	if got.Timestamp != want {
		t.Errorf("Timestamp = %#x, want %#x", got.Timestamp, want)
	}
}

func TestStatusRequestServerboundDecode(t *testing.T) {
	var got StatusRequestServerbound
	r := protocol.NewReader(nil)
	got.Decode(r)

	if err := r.Err(); err != nil {
		t.Fatalf("Decoding an empty body: an unexpected error: %v", err)
	}
}

func TestStatusResponseClientboundEncode(t *testing.T) {
	const json = `{"version":{"name":"26.2","protocol":999},"description":{"text":"hi"}}`
	p := StatusResponseClientbound{JSONResponse: json}

	if p.ID() != 0x00 {
		t.Errorf("ID() = %#x, want 0x00", p.ID())
	}

	w := &protocol.Writer{}
	p.Encode(w)

	r := protocol.NewReader(w.Bytes())
	got := r.String()
	if err := r.Err(); err != nil {
		t.Fatalf("re-decode: error: %v", err)
	}
	if got != json {
		t.Errorf("JSON round-trip = %q, want %q", got, json)
	}
}

func TestPongResponseClientboundEncode(t *testing.T) {
	const want = int64(-42)

	p := PongResponseClientbound{Timestamp: want}
	if p.ID() != 0x01 {
		t.Errorf("ID() = %#x, want 0x01", p.ID())
	}

	w := &protocol.Writer{}
	p.Encode(w)

	r := protocol.NewReader(w.Bytes())
	got := r.I64()
	if err := r.Err(); err != nil {
		t.Fatalf("re-decode: error: %v", err)
	}
	if got != want {
		t.Errorf("Timestamp round-trip = %d, want %d", got, want)
	}
}
