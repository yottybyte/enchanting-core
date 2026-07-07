package network

import (
	"testing"

	"github.com/yottybyte/enchanting-core/internal/adapter/protocol"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func clientFrame(id int32, fields []byte) []byte {
	inner := &protocol.Writer{}
	inner.VarInt(id)
	payload := append(inner.Bytes(), fields...)

	out := &protocol.Writer{}
	out.VarInt(int32(len(payload)))
	return append(out.Bytes(), payload...)
}
