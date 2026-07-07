package protocol

import (
	"bufio"
	"fmt"
	"io"
)

const MaxPacketSize = 2_097_151

func ReadFrame(br *bufio.Reader) (id int32, body *Reader, err error) {
	length, err := ReadVarInt(br)
	if err != nil {
		return id, nil, err
	}

	if length < 0 || length > MaxPacketSize {
		return id, nil, fmt.Errorf("bad frame length: %d", length)
	}

	buf := make([]byte, length)
	_, err = io.ReadFull(br, buf)
	if err != nil {
		return id, nil, err
	}

	r := NewReader(buf)
	id = r.VarInt()

	if r.Err() != nil {
		return id, nil, r.Err()
	}

	return id, r, nil
}

func WritePacket(w io.Writer, pkt Clientbound) error {
	body := &Writer{}
	body.VarInt(pkt.ID())
	pkt.Encode(body)

	frame := &Writer{}
	frame.VarInt(int32(len(body.Bytes())))

	_, err := w.Write(append(frame.Bytes(), body.Bytes()...))
	return err
}
