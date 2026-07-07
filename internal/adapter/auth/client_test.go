package auth

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHasJoined(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("username"); got != "Notch" {
			t.Errorf("username = %q", got)
		}
		if r.URL.Query().Get("serverId") == "" {
			t.Error("serverId is empty")
		}
		_, _ = io.WriteString(w, `{"id":"069a79f444e94726a5befca90e38aaf5","name":"Notch","properties":[{"name":"textures","value":"abc","signature":"sig"}]}`)
	}))
	defer srv.Close()

	c := NewClient()
	c.baseURL = srv.URL

	p, err := c.HasJoined(context.Background(), "", "Notch", make([]byte, 16), []byte{0x01})
	if err != nil {
		t.Fatalf("HasJoined: %v", err)
	}
	if p.ID != "069a79f444e94726a5befca90e38aaf5" || p.Name != "Notch" {
		t.Errorf("profile = %+v", p)
	}
	if len(p.Properties) != 1 || p.Properties[0].Name != "textures" {
		t.Errorf("properties = %+v", p.Properties)
	}
}

func TestHasJoinedNotAuthenticated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := NewClient()
	c.baseURL = srv.URL

	if _, err := c.HasJoined(context.Background(), "", "Notch", make([]byte, 16), []byte{0x01}); !errors.Is(err, ErrNotAuthenticated) {
		t.Fatalf("err = %v, want ErrNotAuthenticated", err)
	}
}
