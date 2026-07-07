package protocol

import (
	"errors"
	"io"
)

// 0x7F = (0111 1111)
// 0x80 = (1000 0000)

func WriteVarInt(w io.ByteWriter, v int32) error {
	u := uint32(v)
	for {
		if u < 0x80 {
			return w.WriteByte(byte(u))
		}
		if err := w.WriteByte(byte(u&0x7F | 0x80)); err != nil {
			return err
		}
		u >>= 7
	}
}

func ReadVarInt(r io.ByteReader) (int32, error) {
	var result uint32
	var shift uint
	for i := 0; i < 5; i++ {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}

		result |= uint32(b&0x7F) << shift
		if b&0x80 == 0 {
			return int32(result), nil
		}
		shift += 7
	}
	return 0, errors.New("VarInt too big")
}

func WriteVarLong(w io.ByteWriter, v int64) error {
	u := uint64(v)
	for {
		if u < 0x80 {
			return w.WriteByte(byte(u))
		}
		if err := w.WriteByte(byte(u&0x7F | 0x80)); err != nil {
			return err
		}
		u >>= 7
	}
}

func ReadVarLong(r io.ByteReader) (int64, error) {
	var result uint64
	var shift uint
	for i := 0; i < 10; i++ {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}

		result |= uint64(b&0x7F) << shift
		if b&0x80 == 0 {
			return int64(result), nil
		}
		shift += 7
	}
	return 0, errors.New("VarLong too big")
}
