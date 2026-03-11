package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

var ErrUnauthenticated = fmt.Errorf("unauthenticated")

type apiErrorResponse struct {
	Error struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"request_id"`
	} `json:"error"`
}

func New() (*Client, error) {
	base := strings.TrimRight(os.Getenv("PROMPTS_API_URL"), "/")
	if base == "" {
		base = "https://api.prompts.dev/v1"
	}

	token, _ := readToken()
	return &Client{BaseURL: base, Token: token, HTTPClient: &http.Client{}}, nil
}

func (c *Client) SetToken(token string) {
	c.Token = token
}

func (c *Client) Do(method, path string, body any, authenticated bool) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if authenticated && c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		defer resp.Body.Close()
		var apiErr apiErrorResponse
		_ = json.NewDecoder(resp.Body).Decode(&apiErr)
		if apiErr.Error.RequestID != "" {
			return nil, fmt.Errorf("%w: request_id=%s", ErrUnauthenticated, apiErr.Error.RequestID)
		}
		return nil, ErrUnauthenticated
	}

	return resp, nil
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".prompts", "config.json")
}

func readToken() (string, error) {
	v := viper.New()
	v.SetConfigFile(configPath())
	v.SetConfigType("json")
	if err := v.ReadInConfig(); err != nil {
		return "", err
	}
	return strings.TrimSpace(v.GetString("token")), nil
}
