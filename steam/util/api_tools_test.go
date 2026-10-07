package util

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/olexandrf/steam-market-web-api/steam/market"
	"github.com/olexandrf/steam-market-web-api/steam/web"
)

func TestBuildURL(t *testing.T) {
	appIdParam := "?appid="
	tests := []struct {
		name     string
		base     string
		apiPath  string
		params   url.Values
		expected string
	}{
		{
			name:     "build steam web api news url",
			base:     web.SteamURL,
			apiPath:  web.SteamNewsURL,
			params:   url.Values{},
			expected: web.SteamURL + web.SteamNewsURL,
		},
		{
			name:    "build steam web api achievements url",
			base:    web.SteamURL,
			apiPath: web.SteamAchievementsURL,
			params: url.Values{
				"appid": {web.CS2ID},
			},
			expected: web.SteamURL + web.SteamAchievementsURL + appIdParam + web.CS2ID,
		},
		{
			name:    "build steam market api prices url",
			base:    market.SteamMarketURL,
			apiPath: market.SteamPriceURL,
			params: url.Values{
				"appid": {web.CS2ID},
			},
			expected: market.SteamMarketURL + market.SteamPriceURL + appIdParam + web.CS2ID,
		},
		{
			name:     "build steam market api listings url",
			base:     market.SteamMarketURL,
			apiPath:  market.SteamMarketListURL,
			params:   url.Values{},
			expected: market.SteamMarketURL + market.SteamMarketListURL,
		},
		{
			name:    "base URL with API path and query parameters",
			base:    "https://example.com",
			apiPath: "/api/v1/items",
			params: url.Values{
				"page":  {"1"},
				"limit": {"20"},
			},
			expected: "https://example.com/api/v1/items?limit=20&page=1",
		},
		{
			name:    "base URL without API path",
			base:    "https://example.com",
			apiPath: "",
			params: url.Values{
				"foo": {"bar"},
			},
			expected: "https://example.com?foo=bar",
		},
		{
			name:     "base URL and API path without parameters",
			base:     "https://example.com",
			apiPath:  "/api/v1/items",
			params:   url.Values{},
			expected: "https://example.com/api/v1/items",
		},
		{
			name:     "base URL with existing path",
			base:     "https://example.com/api",
			apiPath:  "v1/items",
			params:   url.Values{},
			expected: "https://example.com/api/v1/items",
		},
		{
			name:    "query parameters are URL encoded",
			base:    "https://example.com",
			apiPath: "/search",
			params: url.Values{
				"query": {"hello world"},
				"tag":   {"go&golang"},
			},
			expected: "https://example.com/search?query=hello+world&tag=go%26golang",
		},
		{
			name:    "multiple values for the same parameter",
			base:    "https://example.com",
			apiPath: "/items",
			params: url.Values{
				"id": {"1", "2", "3"},
			},
			expected: "https://example.com/items?id=1&id=2&id=3",
		},
		{
			name:     "trailing slash in base URL",
			base:     "https://example.com/",
			apiPath:  "/api/items",
			params:   url.Values{},
			expected: "https://example.com/api/items",
		},
		{
			name:     "empty base URL",
			base:     "",
			apiPath:  "/api/items",
			params:   url.Values{},
			expected: "api/items",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildURL(tt.base, tt.apiPath, tt.params)
			if err != nil {
				t.Fatalf("BuildURL() returned unexpected error: %v", err)
			}

			if got != tt.expected {
				t.Errorf(
					"BuildURL() = %q, want %q",
					got,
					tt.expected,
				)
			}
		})
	}
}

func TestBuildURL_InvalidBaseURL(t *testing.T) {
	_, err := BuildURL("://invalid-url", "/api", nil)
	if err == nil {
		t.Fatal("BuildURL() expected an error, got nil")
	}
	if !strings.Contains(err.Error(), "parse url:") {
		t.Fatalf("BuildURL() error = %v, want wrap parse url", err)
	}
}

func TestDoRequest_NilClient(t *testing.T) {
	_, err := DoRequest(t.Context(), nil, "https://go.dev/")
	if !errors.Is(err, ErrClientNil) {
		t.Fatalf("DoJSON() error = %v, want %v", err, ErrClientNil)
	}
}

func TestRequest_MalformedURL(t *testing.T) {
	_, err := DoRequest(t.Context(), &http.Client{}, "://invalid-url")
	if err == nil {
		t.Fatal("DoRequest() error = nil, want create request error")
	}
}

func TestDoRequest_RequestTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()

	_, err := DoRequest(ctx, srv.Client(), srv.URL)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("DoRequest() error = %v, want %v", err, context.DeadlineExceeded)
	}
}

func TestDoRequest_NonOKStatus(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{
			name:       "429 too many requests",
			statusCode: http.StatusTooManyRequests,
		},
		{
			name:       "500 response",
			statusCode: http.StatusInternalServerError,
		},
		{
			name:       "404 response",
			statusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "", tt.statusCode)
			}))
			defer srv.Close()
			client := srv.Client()
			_, err := DoRequest(context.Background(), client, srv.URL)
			if err == nil {
				t.Fatalf("DoRequest() error = nil, want non-nil")
			}
			apiErr, ok := errors.AsType[*APIError](err)
			if !ok {
				t.Fatalf("DoRequest() error = %T(%v), want APIError", apiErr, err)
			}
			if apiErr.StatusCode != tt.statusCode {
				t.Fatalf("DoRequest() error = %v, want %v", apiErr.StatusCode, tt.statusCode)
			}
		})
	}
}

func TestDoRequest_ContextDeadline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := DoRequest(ctx, srv.Client(), srv.URL)
	if err == nil {
		t.Fatal("expected error for deadline context")
	}

	ok := errors.Is(err, context.DeadlineExceeded)
	if !ok {
		t.Fatalf("DoRequest() error = %T(%v), want DeadlineExceeded", err, err)
	}
}

func TestDoRequest_ClientTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()
	client := srv.Client()
	client.Timeout = 50 * time.Millisecond
	_, err := DoRequest(context.Background(), client, srv.URL)
	if err == nil {
		t.Fatal("expected error for timeout context")
	}
	var netErr net.Error
	if !errors.As(err, &netErr) {
		t.Fatalf("DoRequest() error = %T(%v), want a net.Error", err, err)
	}
	if !netErr.Timeout() {
		t.Fatalf("DoRequest() error = %v, want %v", err, context.DeadlineExceeded)
	}
}
