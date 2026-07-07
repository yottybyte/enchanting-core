package network

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/yottybyte/enchanting-core/internal/adapter/protocol"
	"github.com/yottybyte/enchanting-core/internal/config"
)

func TestServerStatusPing(t *testing.T) {
	cfg := &config.Config{Server: config.Server{
		Port: 0, OnlineMode: false, MaxPlayers: 20, MOTD: "test",
	}}
	srv := NewServer(cfg, NewConnOfflineTransport())
	if err := srv.Listen(); err != nil {
		t.Fatalf("listen: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	serveErr := make(chan error, 1)
	go func() { serveErr <- srv.Serve(ctx) }()

	conn, err := net.Dial("tcp", srv.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))

	hf := &protocol.Writer{}
	hf.VarInt(767)
	hf.String("127.0.0.1")
	hf.U16(25565)
	hf.VarInt(1)
	if _, err := conn.Write(clientFrame(0x00, hf.Bytes())); err != nil {
		t.Fatalf("write handshake: %v", err)
	}

	if _, err := conn.Write(clientFrame(0x00, nil)); err != nil {
		t.Fatalf("write status request: %v", err)
	}

	br := bufio.NewReader(conn)
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

	const ping = int64(0x0123456789ABCDEF)
	pf := &protocol.Writer{}
	pf.I64(ping)
	if _, err := conn.Write(clientFrame(0x01, pf.Bytes())); err != nil {
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

	_ = conn.Close()

	cancel()
	select {
	case err := <-serveErr:
		if err != nil {
			t.Fatalf("serve return error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Serve did not complete after canceling ctx")
	}
}
