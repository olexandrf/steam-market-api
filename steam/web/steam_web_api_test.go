package web

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

func TestSteamWebAPIClient_GetNews(t *testing.T) {
	params := url.Values{}
	count := 1
	params.Set("appid", CS2ID)
	params.Set("format", DefaultFormat)
	params.Set("count", strconv.Itoa(count))
	testTitle := "testing-title-content"
	var gotPath, gotAccept string
	testJson := `{"appnews":{"appid":730,"newsitems":[{"gid":"1","title":"` + testTitle + `","url":"https://steamstore-a.akamaihd.net/news/externalpost/GamingOnLinux/1844751498218029","is_external_url":true,"author":"","contents":"test content","feedlabel":"GamingOnLinux","date":17971779988,"feedname":"GamingOnLinux","feed_type":0,"appid":730}],"count":1500}}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAccept = r.Header.Get("Accept")
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, testJson)
	}))
	defer srv.Close()

	SteamMarketNews, err := NewSteamWebAPIClient(srv.Client(), srv.URL).GetNews(context.Background(), CS2ID, 1, DefaultFormat)
	if gotPath != SteamNewsURL {
		t.Errorf("got path %s; want %s", gotPath, SteamNewsURL)
	}
	if gotAccept != "application/json" {
		t.Errorf("got accept %s; want application/json", gotAccept)
	}
	if err != nil {
		t.Fatalf("SteamWebAPIClient.GetNews() error = %v, want nil", err)
	}
	if SteamMarketNews == nil {
		t.Fatalf("SteamWebAPIClient.GetNews() = %v, want non-nil", SteamMarketNews)
	}
	cs2id, _ := strconv.Atoi(CS2ID)
	appNews := SteamMarketNews.AppNews
	if appNews.Appid != cs2id {
		t.Errorf("SteamWebAPIClient.GetNews() = %v, want appid", CS2ID)
	}
	newsItems := appNews.NewsItems
	if len(newsItems) == 0 {
		t.Errorf("SteamWebAPIClient.GetNews() = %v, want non-empty", appNews.NewsItems)
	}
	if newsItems[0].Title != testTitle {
		t.Errorf("SteamWebAPIClient.GetNews() = %v, want title", testTitle)
	}
}
