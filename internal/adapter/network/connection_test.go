package network

import (
	"bufio"
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/yottybyte/enchanting-core/internal/adapter/protocol"
	"github.com/yottybyte/enchanting-core/internal/config"
)

func TestHandleConn(t *testing.T) {
	client, server := net.Pipe()

	cfg := &config.Config{Server: config.Server{
		Port: 0, OnlineMode: false, MaxPlayers: 20, MOTD: "test",
	}}
	srv := NewServer(cfg, NewConnOfflineTransport())

	done := make(chan struct{})
	go func() {
		srv.handleConn(server)
		close(done)
	}()

	_ = client.SetDeadline(time.Now().Add(2 * time.Second))

	// 1) Handshake (next=1 → Status), no answer
	hf := &protocol.Writer{}
	hf.VarInt(767)
	hf.String("localhost")
	hf.U16(25565)
	hf.VarInt(1)
	if _, err := client.Write(clientFrame(0x00, hf.Bytes())); err != nil {
		t.Fatalf("write handshake: %v", err)
	}

	// 2) Status Request → Status Response
	if _, err := client.Write(clientFrame(0x00, nil)); err != nil {
		t.Fatalf("write status request: %v", err)
	}

	br := bufio.NewReader(client)
	id, body, err := protocol.ReadFrame(br)
	if err != nil {
		t.Fatalf("read status response: %v", err)
	}
	if id != 0x00 {
		t.Errorf("status id = %#x, want 0x00", id)
	}
	if js := body.String(); !json.Valid([]byte(js)) {
		t.Errorf("status JSON invalid: %q", js)
	}

	// 3) Ping → Pong
	const ping = int64(0x0123456789ABCDEF)
	pf := &protocol.Writer{}
	pf.I64(ping)
	if _, err := client.Write(clientFrame(0x01, pf.Bytes())); err != nil {
		t.Fatalf("write ping: %v", err)
	}
	id, body, err = protocol.ReadFrame(br)
	if err != nil {
		t.Fatalf("read pong: %v", err)
	}
	if id != 0x01 {
		t.Errorf("pong id = %#x, want 0x01", id)
	}
	if got := body.I64(); got != ping {
		t.Errorf("pong = %#x, want %#x", got, ping)
	}

	// 4) client exits → handleConn exits via io.EOF
	_ = client.Close()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("handleConn did not terminate after the client was closed")
	}
}
