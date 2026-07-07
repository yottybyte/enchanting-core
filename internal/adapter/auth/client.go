package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/yottybyte/enchanting-core/internal/domain"
)

var ErrNotAuthenticated = errors.New("auth: player not authenticated")

const defaultBaseURL = "https://sessionserver.mojang.com"

type Client struct {
	http    *http.Client
	baseURL string
}

func NewClient() *Client {
	return &Client{
		http:    &http.Client{Timeout: 10 * time.Second},
		baseURL: defaultBaseURL,
	}
}

func (c *Client) HasJoined(ctx context.Context, serverID, username string, sharedSecret, publicKey []byte) (*domain.Profile, error) {
	hash := authDigest(serverID, sharedSecret, publicKey)

	parsedUrl, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}
	parsedUrl.Path = "/session/minecraft/hasJoined"

	queryParams := parsedUrl.Query()
	queryParams.Add("username", username)
	queryParams.Add("serverId", hash)
	parsedUrl.RawQuery = queryParams.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", parsedUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNoContent {
		return nil, ErrNotAuthenticated
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("auth: unexpected status code: %d", resp.StatusCode)
		return nil, errors.New("auth: unexpected status code")
	}

	var p = new(domain.Profile)
	err = json.NewDecoder(resp.Body).Decode(p)
	if err != nil {
		return nil, err
	}

	return p, nil
}
