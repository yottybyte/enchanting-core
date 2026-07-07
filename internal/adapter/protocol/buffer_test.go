package protocol

import (
	"bytes"
	"testing"
)

func TestBufferRoundTrip(t *testing.T) {
	w := &Writer{}
	w.Bool(true)
	w.U16(0x1234)
	w.U32(0xDEADBEEF)
	w.U64(0x1122334455667788)
	w.I16(-1000)
	w.I32(-123456)
	w.I64(-9876543210)
	w.F32(3.14)
	w.F64(2.718281828)
	w.VarInt(300)
	w.VarLong(-1)

	r := NewReader(w.Bytes())
	if got := r.Bool(); !got {
		t.Errorf("Bool = %v, want true", got)
	}
	if got := r.U16(); got != 0x1234 {
		t.Errorf("U16 = %#x, want 0x1234", got)
	}
	if got := r.U32(); got != 0xDEADBEEF {
		t.Errorf("U32 = %#x, want 0xDEADBEEF", got)
	}
	if got := r.U64(); got != 0x1122334455667788 {
		t.Errorf("U64 = %#x, want 0x1122334455667788", got)
	}
	if got := r.I16(); got != -1000 {
		t.Errorf("I16 = %d, want -1000", got)
	}
	if got := r.I32(); got != -123456 {
		t.Errorf("I32 = %d, want -123456", got)
	}
	if got := r.I64(); got != -9876543210 {
		t.Errorf("I64 = %d, want -9876543210", got)
	}
	if got := r.F32(); got != 3.14 {
		t.Errorf("F32 = %v, want 3.14", got)
	}
	if got := r.F64(); got != 2.718281828 {
		t.Errorf("F64 = %v, want 2.718281828", got)
	}
	if got := r.VarInt(); got != 300 {
		t.Errorf("VarInt = %d, want 300", got)
	}
	if got := r.VarLong(); got != -1 {
		t.Errorf("VarLong = %d, want -1", got)
	}
	if err := r.Err(); err != nil {
		t.Errorf("Err() = %v, want nil", err)
	}
}

func TestWriterBigEndian(t *testing.T) {
	w := &Writer{}
	w.U16(0x1234)
	w.U32(0xDEADBEEF)
	want := []byte{0x12, 0x34, 0xDE, 0xAD, 0xBE, 0xEF}
	if !bytes.Equal(w.Bytes(), want) {
		t.Errorf("Bytes = % x, want % x", w.Bytes(), want)
	}
}

func TestReaderTruncated(t *testing.T) {
	r := NewReader([]byte{0x00, 0x01})
	_ = r.U32()
	if r.Err() == nil {
		t.Fatal("Err() = nil, want a byte shortage error")
	}

	if got := r.U64(); got != 0 {
		t.Errorf("U64 after an error = %d, want 0", got)
	}
}
