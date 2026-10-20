package steam

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

const (
	MarketURL      = "https://steamcommunity.com/market"
	MarketListURL  = "/search/render"
	PriceURL       = "/priceoverview"
	SaleHistoryURL = "/listings"
	NoRender       = "1"
)

type MarketClient struct {
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

// PriceHistoryEntry is one aggregated Steam Market sale data point. Price and
// Quantity are strings because Steam returns them as strings and the price can
// use a locale-specific decimal separator.
type PriceHistoryEntry struct {
	Date     string
	Price    string
	Quantity string
}

// SaleHistory contains the data embedded in a market listing page. Steam
// serves this endpoint as HTML, with the useful data assigned to JavaScript
// variables. ListingInfo and Assets retain Steam's original JSON shape so the
// client does not discard item-specific fields.
type SaleHistory struct {
	PriceHistory []PriceHistoryEntry
	ListingInfo  map[string]json.RawMessage
	Assets       map[string]json.RawMessage
}

type PriceOverview struct {
	Success     bool   `json:"success"`
	LowestPrice string `json:"lowest_price"`
	Volume      string `json:"volume"`
	MedianPrice string `json:"median_price"`
}

func NewMarketClient(httpClient *http.Client, url string) *MarketClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: DefaultTimeout}
	}
	return &MarketClient{Client: httpClient, url: url}
}

// GetMarketList returns one page of searchable Steam Market items.
func (s *MarketClient) GetMarketList(ctx context.Context, appid int, query, sortColumn, sortDir string, start, count int) (*ItemsList, error) {
	params := url.Values{}
	params.Set("norender", NoRender)
	params.Set("appid", strconv.Itoa(appid))
	params.Set("query", query)
	params.Set("sortColumn", sortColumn)
	params.Set("sortDir", sortDir)
	params.Set("start", strconv.Itoa(start))
	params.Set("count", strconv.Itoa(count))

	reqURL, err := buildURL(s.url, MarketListURL, params)
	if err != nil {
		return nil, fmt.Errorf("get market list: %w", err)
	}

	result, err := doJSON[ItemsList](ctx, s.Client, reqURL)
	if err != nil {
		return nil, fmt.Errorf("get market list: %w", err)
	}
	if result == nil {
		return nil, fmt.Errorf("get market list: result is nil")
	}

	return result, nil
}

// GetPrices returns the current price overview for a market item.
func (s *MarketClient) GetPrices(ctx context.Context, appid string, currency int, marketHashName string) (*PriceOverview, error) {
	params := url.Values{}
	params.Set("appid", appid)
	params.Set("currency", strconv.Itoa(currency))
	params.Set("market_hash_name", marketHashName)

	reqURL, err := buildURL(s.url, PriceURL, params)
	if err != nil {
		return nil, fmt.Errorf("get prices: %w", err)
	}

	price, err := doJSON[PriceOverview](ctx, s.Client, reqURL)

	if err != nil {
		return nil, fmt.Errorf("get prices: %w", err)
	}

	if price == nil {
		return nil, fmt.Errorf("get prices: result is nil")
	}

	return price, nil
}

// GetSaleHistory returns sale history and listing data from a Steam Market
// listing page. itemName is the market hash name (for example, "AK-47 | Redline
// (Field-Tested)"). Steam returns HTML rather than a JSON API response; the
// embedded price-history, listing, and asset JavaScript variables are parsed.
func (s *MarketClient) GetSaleHistory(ctx context.Context, appid int, itemName, language string, currency, start, count int) (*SaleHistory, error) {
	params := url.Values{}
	params.Set("language", language)
	params.Set("currency", strconv.Itoa(currency))
	params.Set("start", strconv.Itoa(start))
	params.Set("count", strconv.Itoa(count))
	params.Set("norender", NoRender)
	uri, err := url.JoinPath(SaleHistoryURL, url.PathEscape(strconv.Itoa(appid)), url.PathEscape(itemName))
	if err != nil {
		return nil, fmt.Errorf("get sale history: build path: %w", err)
	}
	reqURL, err := buildURL(s.url, uri, params)
	if err != nil {
		return nil, fmt.Errorf("get sale history: %w", err)
	}

	result, err := doRequest(ctx, s.Client, reqURL)
	if err != nil {
		return nil, fmt.Errorf("get sale history: %w", err)
	}

	history, err := parseSaleHistory(result)
	if err != nil {
		return nil, fmt.Errorf("get sale history: %w", err)
	}
	return history, nil
}
