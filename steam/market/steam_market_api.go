package market

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/olexandrf/steam-market-web-api/steam/util"
)

const (
	SteamMarketURL     = "https://steamcommunity.com/market"
	SteamMarketListURL = "/search/render"
	SteamPriceURL      = "/priceoverview"
	noRender           = "1"
)

type SteamMarketClient struct {
	Client *http.Client
	url    string
}

type SearchData struct {
	Query       string `json:"query"`
	TotalCount  int    `json:"total_count"`
	PageSize    int    `json:"pagesize"`
	Prefix      string `json:"prefix"`
	ClassPrefix string `json:"class_prefix"`
}

type AssetDescription struct {
	Appid                 int    `json:"appid"`
	ClassID               string `json:"classid"`
	BackgroundColor       string `json:"background_color"`
	IconURL               string `json:"icon_url"`
	Tradable              int    `json:"tradable"`
	Name                  string `json:"name"`
	NameColor             string `json:"name_color"`
	Type                  string `json:"type"`
	MarketName            string `json:"market_name"`
	MarketHashName        string `json:"market_hash_name"`
	Commodity             int    `json:"commodity"`
	MarketBucketGroupName string `json:"market_bucket_group_name"`
	MarketBucketGroupID   string `json:"market_bucket_group_id"`
}

type ResultItem struct {
	Name             string           `json:"name"`
	HashName         string           `json:"hash_name"`
	SellListings     int              `json:"sell_listings"`
	SellPrice        int              `json:"sell_price"`
	SellPriceText    string           `json:"sell_price_text"`
	AppIcon          string           `json:"app_icon"`
	AppName          string           `json:"app_name"`
	AssetDescription AssetDescription `json:"asset_description"`
	SalePriceText    string           `json:"sale_price_text"`
}

type ItemsList struct {
	Success    bool         `json:"success"`
	Start      int          `json:"start"`
	PageSize   int          `json:"pagesize"`
	TotalCount int          `json:"total_count"`
	SearchData SearchData   `json:"searchdata"`
	Results    []ResultItem `json:"results"`
}

type PriceOverview struct {
	Success     bool   `json:"success"`
	LowestPrice string `json:"lowest_price"`
	Volume      string `json:"volume"`
	MedianPrice string `json:"median_price"`
}

func NewSteamMarketClient(httpClient *http.Client, url string) *SteamMarketClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: util.DefaultTimeout}
	}
	return &SteamMarketClient{Client: httpClient, url: url}
}

func (s *SteamMarketClient) GetMarketList(ctx context.Context, appid string) (*ItemsList, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}
	params := url.Values{}
	params.Set("norender", noRender)
	params.Set("appid", appid)

	reqURL, err := util.BuildURL(s.url, SteamMarketListURL, params)
	if err != nil {
		return nil, fmt.Errorf("get market list: %w", err)
	}

	result, err := util.DoJSON[ItemsList](ctx, s.Client, reqURL)
	if err != nil {
		return nil, fmt.Errorf("get market list: %w", err)
	}
	if result == nil {
		return nil, fmt.Errorf("get market list: result is nil")
	}

	return result, nil
}

func (s *SteamMarketClient) GetPrices(ctx context.Context, appid string, currency string, marketHashName string) (*PriceOverview, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}
	params := url.Values{}
	params.Set("appid", appid)
	params.Set("currency", currency)
	params.Set("market_hash_name", marketHashName)

	reqURL, err := util.BuildURL(s.url, SteamPriceURL, params)
	if err != nil {
		return nil, fmt.Errorf("get prices: %w", err)
	}

	price, err := util.DoJSON[PriceOverview](ctx, s.Client, reqURL)

	if err != nil {
		return nil, fmt.Errorf("get prices: %w", err)
	}

	if price == nil {
		return nil, fmt.Errorf("get prices: result is nil")
	}

	return price, nil
}
