package steam

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSteamMarketClient_GetPrices(t *testing.T) {
	var gotURL, gotAccept string
	testJson := `{"success":true,"lowest_price":"$33.33","volume":"42","median_price":"$40.40"}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.String()
		gotAccept = r.Header.Get("Accept")
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, testJson)
	}))

	defer srv.Close()
	currency := "1"
	price, err := NewMarketClient(srv.Client(), srv.URL).GetPrices(context.Background(), CS2ID, currency, "test_hash_name")
	if gotURL != "/priceoverview"+"?appid=730&currency=1&market_hash_name=test_hash_name" {
		t.Fatalf("GetPrices: got URL %s, want %s", gotURL, PriceURL)
	}
	if gotAccept != "application/json" {
		t.Fatalf("GetPrices: got Accept %s, want %s", gotAccept, "application/json")
	}
	if err != nil {
		t.Fatalf("GetPrices: got err %v, want nil", err)
	}
	if price == nil {
		t.Fatal("GetPrices: got nil, want non-nil")
	}
	if price.Success != true && price.LowestPrice != "$33.33" && price.LowestPrice != "$40.40" {
		t.Fatalf("GetPrices: got Price %s, want %s", price.LowestPrice, currency)
	}
}
