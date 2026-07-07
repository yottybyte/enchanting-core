package protocol

import (
	"bufio"
	"bytes"
	"io"
	"testing"
)

type framePkt struct {
	id  int32
	msg string
}

func (p *framePkt) ID() int32        { return p.id }
func (p *framePkt) Encode(w *Writer) { w.String(p.msg) }

var _ Clientbound = (*framePkt)(nil)

func TestFrameRoundTrip(t *testing.T) {
	cases := []framePkt{
		{id: 0x00, msg: "hello frame"},
		{id: 0x01, msg: ""},
		{id: 300, msg: "multibyte ID and Unicode 🧱"},
		{id: 127, msg: "x"},
	}
	for i := range cases {
		pkt := &cases[i]

		var buf bytes.Buffer
		if err := WritePacket(&buf, pkt); err != nil {
			t.Fatalf("WritePacket(%+v): %v", pkt, err)
		}

		id, body, err := ReadFrame(bufio.NewReader(&buf))
		if err != nil {
			t.Fatalf("ReadFrame(%+v): %v", pkt, err)
		}
		if id != pkt.id {
			t.Errorf("id = %d, want %d", id, pkt.id)
		}
		got := body.String()
		if err := body.Err(); err != nil {
			t.Fatalf("body decode(%+v): %v", pkt, err)
		}
		if got != pkt.msg {
			t.Errorf("payload = %q, want %q", got, pkt.msg)
		}
	}
}

func TestReadFrameEOF(t *testing.T) {
	_, _, err := ReadFrame(bufio.NewReader(bytes.NewReader(nil)))
	if err != io.EOF {
		t.Fatalf("we expected io.EOF on an empty stream, but received %v", err)
	}
}

func TestReadFrameTruncated(t *testing.T) {
	var buf bytes.Buffer
	w := &Writer{}
	w.VarInt(10)
	buf.Write(w.Bytes())
	buf.WriteString("abc")

	_, _, err := ReadFrame(bufio.NewReader(&buf))
	if err == nil {
		t.Fatal("we expected an error on the cropped frame, but got nil")
	}
}

func TestReadFrameTooLarge(t *testing.T) {
	var buf bytes.Buffer
	w := &Writer{}
	w.VarInt(MaxPacketSize + 1)
	buf.Write(w.Bytes())

	_, _, err := ReadFrame(bufio.NewReader(&buf))
	if err == nil {
		t.Fatal("we expected a limit exceeded error, but got nil")
	}
}
