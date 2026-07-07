package protocol

import (
	"bytes"
	"testing"
)

func TestStringRoundTrip(t *testing.T) {
	for _, s := range []string{"", "minecraft:overworld", "steve 🧱"} {
		w := &Writer{}
		w.String(s)
		got := NewReader(w.Bytes()).String()
		if got != s {
			t.Errorf("String round-trip %q = %q", s, got)
		}
	}
}

func TestStringEncoding(t *testing.T) {
	w := &Writer{}
	w.String("abc")
	want := []byte{0x03, 'a', 'b', 'c'}
	if !bytes.Equal(w.Bytes(), want) {
		t.Errorf("String(\"abc\") = % x, want % x", w.Bytes(), want)
	}
	w2 := &Writer{}
	w2.String("é")
	if w2.Bytes()[0] != 0x02 {
		t.Errorf("len prefix for \"é\" = %d, want 2 (byte)", w2.Bytes()[0])
	}
}

func TestStringErrors(t *testing.T) {
	neg := &Writer{}
	neg.VarInt(-1)
	r := NewReader(neg.Bytes())
	_ = r.String()
	if r.Err() == nil {
		t.Error("String(negative length): want error")
	}

	r2 := NewReader([]byte{0x0A, 'a', 'b', 'c'})
	_ = r2.String()
	if r2.Err() == nil {
		t.Error("String(truncated): want error")
	}
}

func TestUUIDRoundTrip(t *testing.T) {
	u := UUID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	w := &Writer{}
	w.UUID(u)
	if len(w.Bytes()) != 16 {
		t.Fatalf("UUID → %d byte, want 16", len(w.Bytes()))
	}
	if got := NewReader(w.Bytes()).UUID(); got != u {
		t.Errorf("UUID = %v, want %v", got, u)
	}
}

func TestPositionRoundTrip(t *testing.T) {
	cases := []Position{
		{0, 0, 0},
		{1, 2, 3},
		{-1, -1, -1},
		{-33554432, -2048, 33554431},
		{33554431, 2047, -33554432},
	}
	for _, p := range cases {
		w := &Writer{}
		w.Position(p)
		if got := NewReader(w.Bytes()).Position(); got != p {
			t.Errorf("Position round-trip %+v = %+v", p, got)
		}
	}
}

func TestByteArrayRoundTrip(t *testing.T) {
	cases := [][]byte{
		nil,
		{},
		{0x00},
		{0xde, 0xad, 0xbe, 0xef},
		bytes.Repeat([]byte{0xAB}, 300),
	}
	for _, want := range cases {
		w := &Writer{}
		w.ByteArray(want)

		r := NewReader(w.Bytes())
		got := r.ByteArray()
		if err := r.Err(); err != nil {
			t.Fatalf("len=%d: unexpected error: %v", len(want), err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("len=%d: round-trip = %x, want %x", len(want), got, want)
		}
	}
}

func TestByteArrayTruncated(t *testing.T) {
	r := NewReader([]byte{0x05, 0x01, 0x02})
	_ = r.ByteArray()
	if r.Err() == nil {
		t.Fatal("expected error on truncated byte array, got nil")
	}
}

func TestByteArrayNegativeLength(t *testing.T) {
	w := &Writer{}
	w.VarInt(-1)
	r := NewReader(w.Bytes())
	_ = r.ByteArray()
	if r.Err() == nil {
		t.Fatal("expected error on negative length, got nil")
	}
}

func TestParseUUID(t *testing.T) {
	want := UUID{0x7d, 0x6d, 0x2e, 0x7a, 0xc6, 0x33, 0x4a, 0xb5, 0x97, 0x20, 0x84, 0x80, 0xf4, 0x0b, 0x14, 0x5b}
	for _, in := range []string{
		"7d6d2e7ac6334ab597208480f40b145b",
		"7d6d2e7a-c633-4ab5-9720-8480f40b145b",
	} {
		got, err := ParseUUID(in)
		if err != nil {
			t.Fatalf("ParseUUID(%q): %v", in, err)
		}
		if got != want {
			t.Errorf("ParseUUID(%q) = %x, want %x", in, got, want)
		}
	}
}
