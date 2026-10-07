package util

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	MaxErrorBody   = 1 << 10 // 1024
	DefaultTimeout = 5 * time.Second
)

var ErrClientNil = errors.New("client is nil")

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("http %d", e.StatusCode)
	}
	return fmt.Sprintf("http %d: %s", e.StatusCode, e.Message)
}

func BuildURL(base string, apiPath string, params url.Values) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}
	if apiPath != "" {
		u = u.JoinPath(apiPath)
	}
	u.RawQuery = params.Encode()
	return u.String(), nil
}

func DoRequest(ctx context.Context, client *http.Client, rawURL string) ([]byte, error) {
	if client == nil {
		return nil, ErrClientNil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, MaxErrorBody))
		return nil, &APIError{StatusCode: resp.StatusCode, Message: string(body)}
	}

	return io.ReadAll(resp.Body)
}

func DoJSON[T any](ctx context.Context, client *http.Client, rawURL string) (*T, error) {
	respBody, err := DoRequest(ctx, client, rawURL)
	if err != nil {
		return nil, err
	}

	return ParseJSONResponse[T](respBody)
}

func ParseJSONResponse[T any](body []byte) (*T, error) {
	var v T
	if err := json.Unmarshal(body, &v); err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}

	return &v, nil
}
