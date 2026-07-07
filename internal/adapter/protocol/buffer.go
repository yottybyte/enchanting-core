package protocol

import (
	"encoding/binary"
	"io"
	"math"
)

type Reader struct {
	buf []byte
	pos int
	err error
}

func NewReader(buf []byte) *Reader {
	return &Reader{buf: buf}
}
func (r *Reader) Err() error {
	return r.err
}

func (r *Reader) next(n int) []byte {
	if r.err != nil {
		return nil
	}
	if r.pos+n > len(r.buf) {
		r.err = io.ErrUnexpectedEOF
		return nil
	}
	b := r.buf[r.pos : r.pos+n]
	r.pos += n
	return b
}

func (r *Reader) ReadByte() (byte, error) {
	b := r.next(1)
	if b == nil {
		return 0, r.err
	}
	return b[0], nil
}

func (r *Reader) Bool() bool {
	b := r.next(1)
	if b == nil {
		return false
	}
	return b[0] != 0
}

func (r *Reader) U16() uint16 {
	b := r.next(2)
	if b == nil {
		return 0
	}
	return binary.BigEndian.Uint16(b)
}
func (r *Reader) U32() uint32 {
	b := r.next(4)
	if b == nil {
		return 0
	}
	return binary.BigEndian.Uint32(b)
}
func (r *Reader) U64() uint64 {
	b := r.next(8)
	if b == nil {
		return 0
	}
	return binary.BigEndian.Uint64(b)
}

func (r *Reader) I16() int16 {
	b := r.next(2)
	if b == nil {
		return 0
	}
	return int16(binary.BigEndian.Uint16(b))
}
func (r *Reader) I32() int32 {
	b := r.next(4)
	if b == nil {
		return 0
	}
	return int32(binary.BigEndian.Uint32(b))
}
func (r *Reader) I64() int64 {
	b := r.next(8)
	if b == nil {
		return 0
	}
	return int64(binary.BigEndian.Uint64(b))
}

func (r *Reader) F32() float32 {
	b := r.next(4)
	if b == nil {
		return 0
	}
	return math.Float32frombits(binary.BigEndian.Uint32(b))
}
func (r *Reader) F64() float64 {
	b := r.next(8)
	if b == nil {
		return 0
	}
	return math.Float64frombits(binary.BigEndian.Uint64(b))
}

func (r *Reader) VarInt() int32 {
	if r.err != nil {
		return 0
	}
	v, err := ReadVarInt(r)
	if err != nil {
		r.err = err
	}
	return v
}
func (r *Reader) VarLong() int64 {
	if r.err != nil {
		return 0
	}
	v, err := ReadVarLong(r)
	if err != nil {
		r.err = err
	}
	return v
}

type Writer struct {
	buf []byte
}

func (w *Writer) Bytes() []byte {
	return w.buf
}
func (w *Writer) WriteByte(b byte) error {
	w.buf = append(w.buf, b)
	return nil
}

func (w *Writer) Bool(b bool) {
	if b {
		w.buf = append(w.buf, 1)
	} else {
		w.buf = append(w.buf, 0)
	}
}

func (w *Writer) U16(v uint16) {
	w.buf = binary.BigEndian.AppendUint16(w.buf, v)
}
func (w *Writer) U32(v uint32) {
	w.buf = binary.BigEndian.AppendUint32(w.buf, v)
}
func (w *Writer) U64(v uint64) {
	w.buf = binary.BigEndian.AppendUint64(w.buf, v)
}

func (w *Writer) I16(v int16) {
	w.buf = binary.BigEndian.AppendUint16(w.buf, uint16(v))
}
func (w *Writer) I32(v int32) {
	w.buf = binary.BigEndian.AppendUint32(w.buf, uint32(v))
}
func (w *Writer) I64(v int64) {
	w.buf = binary.BigEndian.AppendUint64(w.buf, uint64(v))
}

func (w *Writer) F32(v float32) {
	w.buf = binary.BigEndian.AppendUint32(w.buf, math.Float32bits(v))
}
func (w *Writer) F64(v float64) {
	w.buf = binary.BigEndian.AppendUint64(w.buf, math.Float64bits(v))
}

func (w *Writer) VarInt(v int32) {
	_ = WriteVarInt(w, v)
}
func (w *Writer) VarLong(v int64) {
	_ = WriteVarLong(w, v)
}
