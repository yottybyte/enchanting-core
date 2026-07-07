package packets

import (
	"testing"

	"github.com/yottybyte/enchanting-core/internal/adapter/protocol"
)

func TestHandshakeServerboundDecode(t *testing.T) {
	cases := []HandshakeServerbound{
		{ProtocolVersion: 767, ServerAddress: "localhost", ServerPort: 25565, NextState: 1},
		{ProtocolVersion: 0, ServerAddress: "", ServerPort: 0, NextState: 2},
		{ProtocolVersion: 2147483647, ServerAddress: "mc.example.com", ServerPort: 65535, NextState: 2},
		{ProtocolVersion: 1, ServerAddress: "st.example", ServerPort: 25565, NextState: 1},
	}

	for _, want := range cases {
		w := &protocol.Writer{}
		w.VarInt(want.ProtocolVersion)
		w.String(want.ServerAddress)
		w.U16(want.ServerPort)
		w.VarInt(want.NextState)

		var got HandshakeServerbound
		r := protocol.NewReader(w.Bytes())
		got.Decode(r)

		if err := r.Err(); err != nil {
			t.Fatalf("Decode(%+v): error: %v", want, err)
		}
		if got != want {
			t.Errorf("Decode = %+v, want %+v", got, want)
		}
	}
}

func TestHandshakeServerboundTruncated(t *testing.T) {
	w := &protocol.Writer{}
	w.VarInt(767)
	w.String("localhost")

	var got HandshakeServerbound
	r := protocol.NewReader(w.Bytes())
	got.Decode(r)

	if r.Err() == nil {
		t.Error("Decode on a cropped frame: expected an error, got nil")
	}
}
