package protocol

import (
	"bytes"
	"errors"
	"io"
	"math"
	"testing"
)

var varIntVectors = []struct {
	value int32
	bytes []byte
}{
	{0, []byte{0x00}},
	{1, []byte{0x01}},
	{2, []byte{0x02}},
	{127, []byte{0x7f}},
	{128, []byte{0x80, 0x01}},
	{255, []byte{0xff, 0x01}},
	{25565, []byte{0xdd, 0xc7, 0x01}},
	{2097151, []byte{0xff, 0xff, 0x7f}},
	{2147483647, []byte{0xff, 0xff, 0xff, 0xff, 0x07}},
	{-1, []byte{0xff, 0xff, 0xff, 0xff, 0x0f}},
	{-2147483648, []byte{0x80, 0x80, 0x80, 0x80, 0x08}},
}

func TestWriteVarInt(t *testing.T) {
	for _, c := range varIntVectors {
		var b bytes.Buffer
		if err := WriteVarInt(&b, c.value); err != nil {
			t.Fatalf("WriteVarInt(%d): %v", c.value, err)
		}
		if !bytes.Equal(b.Bytes(), c.bytes) {
			t.Errorf("WriteVarInt(%d) = % x, want % x", c.value, b.Bytes(), c.bytes)
		}
	}
}

func TestReadVarInt(t *testing.T) {
	for _, c := range varIntVectors {
		got, err := ReadVarInt(bytes.NewReader(c.bytes))
		if err != nil {
			t.Fatalf("ReadVarInt(% x): %v", c.bytes, err)
		}
		if got != c.value {
			t.Errorf("ReadVarInt(% x) = %d, want %d", c.bytes, got, c.value)
		}
	}
}

func TestReadVarIntStopsAtBoundary(t *testing.T) {
	r := bytes.NewReader([]byte{0xac, 0x02, 0x2a})
	if v, err := ReadVarInt(r); err != nil || v != 300 {
		t.Fatalf("first = %d, %v; want 300", v, err)
	}
	if v, err := ReadVarInt(r); err != nil || v != 42 {
		t.Fatalf("second = %d, %v; want 42", v, err)
	}
}

func TestReadVarIntErrors(t *testing.T) {

	tooLong := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0x01}
	if _, err := ReadVarInt(bytes.NewReader(tooLong)); err == nil {
		t.Error("ReadVarInt(too long): want error, got nil")
	}

	if _, err := ReadVarInt(bytes.NewReader([]byte{0x80})); !errors.Is(err, io.EOF) {
		t.Errorf("ReadVarInt(truncated): want io.EOF, got %v", err)
	}
}

func TestVarIntRoundTrip(t *testing.T) {
	values := []int32{0, 1, -1, 127, 128, 255, 25565, math.MaxInt32, math.MinInt32}
	for _, v := range values {
		var b bytes.Buffer
		if err := WriteVarInt(&b, v); err != nil {
			t.Fatalf("write %d: %v", v, err)
		}
		got, err := ReadVarInt(&b)
		if err != nil {
			t.Fatalf("read %d: %v", v, err)
		}
		if got != v {
			t.Errorf("round-trip %d = %d", v, got)
		}
	}
}

func FuzzVarIntRoundTrip(f *testing.F) {
	f.Add(int32(0))
	f.Add(int32(-1))
	f.Add(int32(math.MaxInt32))
	f.Fuzz(func(t *testing.T, v int32) {
		var b bytes.Buffer
		if err := WriteVarInt(&b, v); err != nil {
			t.Fatal(err)
		}
		got, err := ReadVarInt(&b)
		if err != nil {
			t.Fatal(err)
		}
		if got != v {
			t.Errorf("round-trip %d = %d", v, got)
		}
	})
}

var varLongVectors = []struct {
	value int64
	bytes []byte
}{
	{0, []byte{0x00}},
	{1, []byte{0x01}},
	{127, []byte{0x7f}},
	{128, []byte{0x80, 0x01}},
	{255, []byte{0xff, 0x01}},
	{2147483647, []byte{0xff, 0xff, 0xff, 0xff, 0x07}},
	{math.MaxInt64, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}},
	{-1, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01}},
	{math.MinInt64, []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x01}},
}

func TestWriteVarLong(t *testing.T) {
	for _, c := range varLongVectors {
		var b bytes.Buffer
		if err := WriteVarLong(&b, c.value); err != nil {
			t.Fatalf("WriteVarLong(%d): %v", c.value, err)
		}
		if !bytes.Equal(b.Bytes(), c.bytes) {
			t.Errorf("WriteVarLong(%d) = % x, want % x", c.value, b.Bytes(), c.bytes)
		}
	}
}

func TestReadVarLong(t *testing.T) {
	for _, c := range varLongVectors {
		got, err := ReadVarLong(bytes.NewReader(c.bytes))
		if err != nil {
			t.Fatalf("ReadVarLong(% x): %v", c.bytes, err)
		}
		if got != c.value {
			t.Errorf("ReadVarLong(% x) = %d, want %d", c.bytes, got, c.value)
		}
	}
}

func TestVarLongRoundTrip(t *testing.T) {
	values := []int64{0, 1, -1, 127, 128, math.MaxInt32, math.MinInt32, math.MaxInt64, math.MinInt64}
	for _, v := range values {
		var b bytes.Buffer
		if err := WriteVarLong(&b, v); err != nil {
			t.Fatalf("write %d: %v", v, err)
		}
		got, err := ReadVarLong(&b)
		if err != nil {
			t.Fatalf("read %d: %v", v, err)
		}
		if got != v {
			t.Errorf("round-trip %d = %d", v, got)
		}
	}
}

func FuzzVarLongRoundTrip(f *testing.F) {
	f.Add(int64(0))
	f.Add(int64(-1))
	f.Add(int64(math.MaxInt64))
	f.Add(int64(math.MinInt64))
	f.Fuzz(func(t *testing.T, v int64) {
		var b bytes.Buffer
		if err := WriteVarLong(&b, v); err != nil {
			t.Fatal(err)
		}
		got, err := ReadVarLong(&b)
		if err != nil {
			t.Fatal(err)
		}
		if got != v {
			t.Errorf("round-trip %d = %d", v, got)
		}
	})
}
