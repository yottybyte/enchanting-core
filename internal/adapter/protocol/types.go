package protocol

import (
	"encoding/hex"
	"errors"
	"strings"
)

func (w *Writer) String(s string) {
	w.VarInt(int32(len(s)))
	w.buf = append(w.buf, s...)
}

func (r *Reader) String() string {
	n := r.VarInt()
	if r.err != nil {
		return ""
	}
	if n < 0 {
		r.err = errors.New("negative string length")
		return ""
	}
	b := r.next(int(n))
	if b == nil {
		return ""
	}
	return string(b)
}

type UUID [16]byte

func (w *Writer) UUID(u UUID) {
	w.buf = append(w.buf, u[:]...)
}

func (r *Reader) UUID() UUID {
	b := r.next(16)
	if b == nil {
		return UUID{}
	}
	return UUID(b)
}

type Position struct{ X, Y, Z int32 }

func (w *Writer) Position(p Position) {
	v := (int64(p.X)&0x3FFFFFF)<<38 | (int64(p.Z)&0x3FFFFFF)<<12 | (int64(p.Y) & 0xFFF)
	w.I64(v)
}

func (r *Reader) Position() Position {
	v := r.I64()
	return Position{
		X: int32(v >> 38),
		Y: int32(v << 52 >> 52),
		Z: int32(v << 26 >> 38),
	}
}

type ByteArray []byte

func (w *Writer) ByteArray(b ByteArray) {
	w.VarInt(int32(len(b)))
	w.buf = append(w.buf, b...)
}

func (r *Reader) ByteArray() ByteArray {
	n := r.VarInt()
	if r.err != nil {
		return nil
	}
	if n < 0 {
		r.err = errors.New("negative byte array length")
		return nil
	}
	b := r.next(int(n))
	if b == nil {
		return nil
	}
	return b
}

func ParseUUID(s string) (UUID, error) {
	b, err := hex.DecodeString(strings.ReplaceAll(s, "-", ""))
	if err != nil {
		return [16]byte{}, err
	}

	if len(b) != 16 {
		return [16]byte{}, errors.New("invalid UUID")
	}

	return UUID(b), nil
}
