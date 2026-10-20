package steam

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestMarketClient_GetPricesTimeoutByContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := NewMarketClient(srv.URL, srv.Client()).GetPrices(ctx, "testid", 1, "test")
	if err == nil {
		t.Fatalf("expected timeout error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded error")
	}
}

func TestMarketClient_GetMarketList(t *testing.T) {
	srv := jsonServer(t, `{"success":true,"total_count":2,"results":[{"name":"AK-47"},{"name":"M4A1-S"}]}`, func(r *http.Request) {
		if r.URL.Path != MarketListURL {
			t.Errorf("path = %q, want %q", r.URL.Path, MarketListURL)
		}
		got := r.URL.Query()
		if got.Get("appid") != "730" || got.Get("query") != "rifle" || got.Get("count") != "10" {
			t.Errorf("unexpected query: %s", got)
		}
	})
	items, err := NewMarketClient(srv.URL, srv.Client()).GetMarketList(context.Background(), CS2ID, "rifle", "name", "asc", 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if !items.Success || items.TotalCount != 2 || len(items.Results) != 2 {
		t.Fatalf("unexpected result: %+v", items)
	}
}

func TestMarketClient_GetPrices(t *testing.T) {
	srv := jsonServer(t, `{"success":true,"lowest_price":"$1.23","volume":"42","median_price":"$1.50"}`, func(r *http.Request) {
		if r.URL.Path != PriceURL {
			t.Errorf("path = %q, want %q", r.URL.Path, PriceURL)
		}
		got := r.URL.Query()
		if got.Get("appid") != "730" || got.Get("currency") != "1" || got.Get("market_hash_name") != "AK-47 | Redline" {
			t.Errorf("unexpected query: %s", got)
		}
	})
	price, err := NewMarketClient(srv.URL, srv.Client()).GetPrices(context.Background(), "730", 1, "AK-47 | Redline")
	if err != nil {
		t.Fatal(err)
	}
	if !price.Success || price.LowestPrice != "$1.23" || price.Volume != "42" {
		t.Fatalf("unexpected result: %+v", price)
	}
}

func TestMarketClient_GetSaleHistory(t *testing.T) {
	body := `<html><script>
var g_rgPriceHistory = [["Jan 01 2025 01:00:00", "1.23", "4"], ["Jan 02", "1,50", "2"]];
var g_rgListingInfo = {"123":{"listingid":"123","asset":{"id":"456"}}};
var g_rgAssets = {"730":{"2":{"456":{"name":"AK-47 | Redline"}}}};
</script></html>`
	srv := jsonServer(t, body, func(r *http.Request) {
		if r.URL.Path != "/listings/730/AK-47 | Redline" {
			t.Errorf("path = %q", r.URL.Path)
		}
		got := r.URL.Query()
		if got.Get("language") != "english" || got.Get("currency") != "1" || got.Get("norender") != "1" {
			t.Errorf("unexpected query: %s", got)
		}
	})
	history, err := NewMarketClient(srv.URL, srv.Client()).GetSaleHistory(context.Background(), CS2ID, "AK-47 | Redline", "english", 1, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(history.PriceHistory) != 2 || history.PriceHistory[1].Price != "1,50" || history.ListingInfo["123"] == nil || history.Assets["730"] == nil {
		t.Fatalf("unexpected history: %+v", history)
	}
}

func TestMarketClient_RequestErrors(t *testing.T) {
	for _, status := range []int{http.StatusBadRequest, http.StatusTooManyRequests, http.StatusInternalServerError} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
				_, _ = w.Write([]byte("unavailable"))
			}))
			defer srv.Close()
			_, err := NewMarketClient(srv.URL, srv.Client()).GetPrices(context.Background(), "730", 1, "item")
			var apiErr *APIError
			if !errors.As(err, &apiErr) || apiErr.StatusCode != status {
				t.Fatalf("error = %v, want APIError status %d", err, status)
			}
		})
	}
}

func TestMarketClient_InvalidResponses(t *testing.T) {
	for _, test := range []struct {
		name, body string
		call       func(*MarketClient) error
	}{
		{"invalid price JSON", `{"success":`, func(c *MarketClient) error { _, err := c.GetPrices(context.Background(), "730", 1, "item"); return err }},
		{"missing sale history", `<script>var g_rgListingInfo = {};</script>`, func(c *MarketClient) error {
			_, err := c.GetSaleHistory(context.Background(), CS2ID, "item", "english", 1, 0, 10)
			return err
		}},
		{"malformed sale history", `<script>var g_rgPriceHistory = [["date", "1", "2"];</script>`, func(c *MarketClient) error {
			_, err := c.GetSaleHistory(context.Background(), CS2ID, "item", "english", 1, 0, 10)
			return err
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			srv := jsonServer(t, test.body, nil)
			if err := test.call(NewMarketClient(srv.URL, srv.Client())); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func jsonServer(t *testing.T, body string, assertRequest func(*http.Request)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if assertRequest != nil {
			assertRequest(r)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
}

func TestBuildURL(t *testing.T) {
	raw, err := buildURL("http://example.test/market", "/priceoverview", url.Values{"appid": {"730"}})
	if err != nil || !strings.Contains(raw, "/market/priceoverview?appid=730") {
		t.Fatalf("buildURL() = %q, %v", raw, err)
	}
}
