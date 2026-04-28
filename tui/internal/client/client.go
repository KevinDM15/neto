// Package client provides an HTTP client for the Neto API and Supabase Auth.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/neto-app/neto/tui/internal/config"
)

const defaultTimeout = 10 * time.Second

// Client handles HTTP communication with the Neto API and Supabase Auth.
type Client struct {
	baseURL         string
	supabaseURL     string
	supabaseAnonKey string
	token           string
	http            *http.Client
}

// New creates a new Client from the given configuration.
func New(cfg *config.Config) *Client {
	return &Client{
		baseURL:         cfg.APIURL,
		supabaseURL:     cfg.SupabaseURL,
		supabaseAnonKey: cfg.SupabaseAnonKey,
		token:           cfg.AccessToken,
		http:            &http.Client{Timeout: defaultTimeout},
	}
}

// SetToken updates the JWT token used for authenticated requests.
func (c *Client) SetToken(token string) {
	c.token = token
}

// Login authenticates via Supabase and returns the access and refresh tokens.
func (c *Client) Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error) {
	url := c.supabaseURL + "/auth/v1/token?grant_type=password"

	body, err := json.Marshal(map[string]string{"email": email, "password": password})
	if err != nil {
		return "", "", fmt.Errorf("client: marshal login body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", "", fmt.Errorf("client: build login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.supabaseAnonKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("client: login request: %w", err)
	}
	defer drainAndClose(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var apiErr struct {
			Message string `json:"message"`
			Error   string `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&apiErr)
		msg := apiErr.Message
		if msg == "" {
			msg = apiErr.Error
		}
		if msg == "" {
			msg = resp.Status
		}
		return "", "", fmt.Errorf("client: login failed: %s", msg)
	}

	var result SupabaseTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", fmt.Errorf("client: decode login response: %w", err)
	}
	return result.AccessToken, result.RefreshToken, nil
}

// Chat sends a message to the Neto chat API and returns the response.
func (c *Client) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("client: marshal chat request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("client: build chat request: %w", err)
	}
	c.setAuthHeader(httpReq)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("client: chat request: %w", err)
	}
	defer drainAndClose(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("client: chat returned %s", resp.Status)
	}

	var result ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("client: decode chat response: %w", err)
	}
	return &result, nil
}

// ListAccounts fetches the user's accounts from the API.
func (c *Client) ListAccounts(ctx context.Context) ([]Account, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/accounts", nil)
	if err != nil {
		return nil, fmt.Errorf("client: build accounts request: %w", err)
	}
	c.setAuthHeader(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("client: accounts request: %w", err)
	}
	defer drainAndClose(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("client: accounts returned %s", resp.Status)
	}

	var accounts []Account
	if err := json.NewDecoder(resp.Body).Decode(&accounts); err != nil {
		return nil, fmt.Errorf("client: decode accounts response: %w", err)
	}
	return accounts, nil
}

func (c *Client) setAuthHeader(req *http.Request) {
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
}

// drainAndClose drains the remaining body so the connection can be reused,
// then closes it. Errors are intentionally ignored — this is best-effort cleanup.
func drainAndClose(body io.ReadCloser) {
	_, _ = io.Copy(io.Discard, body)
	_ = body.Close()
}
