package session

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"testing"

	"github.com/yottybyte/enchanting-core/internal/adapter/protocol"
	"github.com/yottybyte/enchanting-core/internal/domain"
)

type fakeAuth struct {
	profile *domain.Profile
}

func (f fakeAuth) HasJoined(ctx context.Context, serverID, username string, sharedSecret, publicKey []byte) (*domain.Profile, error) {
	return f.profile, nil
}

type testConn struct{ *bytes.Buffer }

func (testConn) EnableEncryption([]byte) error { return nil }

const testStatusJSON = `{"version":{"name":"26.2","protocol":776},"players":{"max":20,"online":0,"sample":[]},"description":{"text":"test"}}`

func TestSessionHandshakeStatusPing(t *testing.T) {
	var out bytes.Buffer
	sess := New(testConn{&out}, &Config{
		StatusJSON: testStatusJSON,
		Auth:       fakeAuth{profile: &domain.Profile{ID: "069a79f444e94726a5befca90e38aaf5", Name: "Notch"}},
	})

	hb := &protocol.Writer{}
	hb.VarInt(767)
	hb.String("localhost")
	hb.U16(25565)
	hb.VarInt(1)
	if err := sess.Handle(0x00, protocol.NewReader(hb.Bytes())); err != nil {
		t.Fatalf("handshake: %v", err)
	}
	if sess.state != protocol.StateStatus {
		t.Fatalf("state = %d, want StateStatus(%d)", sess.state, protocol.StateStatus)
	}
	if out.Len() != 0 {
		t.Errorf("the handshake should not send a response, and there are %d bytes in the stack", out.Len())
	}

	if err := sess.Handle(0x00, protocol.NewReader(nil)); err != nil {
		t.Fatalf("status request: %v", err)
	}

	const ping = int64(0x0123456789ABCDEF)
	pb := &protocol.Writer{}
	pb.I64(ping)
	if err := sess.Handle(0x01, protocol.NewReader(pb.Bytes())); err != nil {
		t.Fatalf("ping: %v", err)
	}

	br := bufio.NewReader(&out)

	id, body, err := protocol.ReadFrame(br)
	if err != nil {
		t.Fatalf("read status response: %v", err)
	}
	if id != 0x00 {
		t.Errorf("status response id = %#x, want 0x00", id)
	}
	js := body.String()
	if err := body.Err(); err != nil {
		t.Fatalf("status response body: %v", err)
	}
	if !json.Valid([]byte(js)) {
		t.Errorf("status JSON invalid: %q", js)
	}

	id, body, err = protocol.ReadFrame(br)
	if err != nil {
		t.Fatalf("read pong: %v", err)
	}
	if id != 0x01 {
		t.Errorf("pong id = %#x, want 0x01", id)
	}
	if got := body.I64(); got != ping {
		t.Errorf("pong payload = %#x, want %#x", got, ping)
	}
	if err := body.Err(); err != nil {
		t.Fatalf("pong body: %v", err)
	}
}

func TestSessionHandshakeToLogin(t *testing.T) {
	var out bytes.Buffer
	sess := New(testConn{&out}, &Config{
		StatusJSON: testStatusJSON,
		Auth:       fakeAuth{profile: &domain.Profile{ID: "069a79f444e94726a5befca90e38aaf5", Name: "Notch"}},
	})

	hb := &protocol.Writer{}
	hb.VarInt(767)
	hb.String("h")
	hb.U16(25565)
	hb.VarInt(2) // next = Login
	if err := sess.Handle(0x00, protocol.NewReader(hb.Bytes())); err != nil {
		t.Fatalf("handshake: %v", err)
	}
	if sess.state != protocol.StateLogin {
		t.Fatalf("state = %d, want StateLogin(%d)", sess.state, protocol.StateLogin)
	}
}

func TestSessionEncryptionResponse(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("genkey: %v", err)
	}

	token := []byte{0x01, 0x02, 0x03, 0x04}
	secret := make([]byte, 16)
	for i := range secret {
		secret[i] = byte(i)
	}

	var out bytes.Buffer
	sess := New(testConn{&out}, &Config{
		PrivateKey: priv,
		StatusJSON: testStatusJSON,
		Auth:       fakeAuth{profile: &domain.Profile{ID: "069a79f444e94726a5befca90e38aaf5", Name: "Notch"}},
	})
	sess.state = protocol.StateLogin
	sess.verifyToken = token

	//nolint:staticcheck //The MC protocol requires PKCS#1 v1.5; OAEP is incompatible with
	encSecret, err := rsa.EncryptPKCS1v15(rand.Reader, &priv.PublicKey, secret)
	if err != nil {
		t.Fatalf("enc secret: %v", err)
	}
	//nolint:staticcheck //The MC protocol requires PKCS#1 v1.5; OAEP is incompatible with
	encToken, err := rsa.EncryptPKCS1v15(rand.Reader, &priv.PublicKey, token)
	if err != nil {
		t.Fatalf("enc token: %v", err)
	}

	body := &protocol.Writer{}
	body.ByteArray(encSecret)
	body.ByteArray(encToken)

	if err := sess.Handle(0x01, protocol.NewReader(body.Bytes())); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if !bytes.Equal(sess.sharedSecret, secret) {
		t.Errorf("sharedSecret = %x, want %x", sess.sharedSecret, secret)
	}
}

func TestSessionEncryptionResponseTokenMismatch(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("genkey: %v", err)
	}

	var out bytes.Buffer
	sess := New(testConn{&out}, &Config{PrivateKey: priv})
	sess.state = protocol.StateLogin
	sess.verifyToken = []byte{0xAA, 0xBB, 0xCC, 0xDD}

	//nolint:staticcheck //The MC protocol requires PKCS#1 v1.5; OAEP is incompatible with
	encSecret, _ := rsa.EncryptPKCS1v15(rand.Reader, &priv.PublicKey, make([]byte, 16))
	//nolint:staticcheck //The MC protocol requires PKCS#1 v1.5; OAEP is incompatible with
	encToken, _ := rsa.EncryptPKCS1v15(rand.Reader, &priv.PublicKey, []byte{0x01, 0x02, 0x03, 0x04})

	body := &protocol.Writer{}
	body.ByteArray(encSecret)
	body.ByteArray(encToken)

	if err := sess.Handle(0x01, protocol.NewReader(body.Bytes())); err == nil {
		t.Fatal("We expected a “verify token mismatch” error, but received nil")
	}
}
