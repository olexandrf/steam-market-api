package web

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/olexandrf/steam-market-web-api/steam/util"
)

const (
	SteamURL             = "https://api.steampowered.com"
	SteamNewsURL         = "/ISteamNews/GetNewsForApp/v0002"
	SteamAchievementsURL = "/ISteamUserStats/GetGlobalAchievementPercentagesForApp/v0002"
	CS2ID                = "730"
	DefaultFormat        = "json"
)

type SteamWebAPIClient struct {
	url string
	c   *http.Client
}

func NewSteamWebAPIClient(client *http.Client, url string) *SteamWebAPIClient {
	if client == nil {
		client = &http.Client{Timeout: util.DefaultTimeout}
	}

	return &SteamWebAPIClient{c: client, url: url}
}

type NewsItem struct {
	Gid           string   `json:"gid"`
	Title         string   `json:"title"`
	URL           string   `json:"url"`
	IsExternalURL bool     `json:"is_external_url"`
	Author        string   `json:"author"`
	Contents      string   `json:"contents"`
	FeedLabel     string   `json:"feedlabel"`
	Date          int      `json:"date"`
	FeedName      string   `json:"feedname"`
	FeedType      int      `json:"feed_type"`
	Appid         int      `json:"appid"`
	Tags          []string `json:"tags,omitempty"`
}

type AppNews struct {
	Appid     int        `json:"appid"`
	NewsItems []NewsItem `json:"newsitems"`
	Count     int        `json:"count"`
}

type SteamNews struct {
	AppNews AppNews `json:"appnews"`
}

type Achievement struct {
	Name    string `json:"name"`
	Percent string `json:"percent"`
}

type SteamAchievements struct {
	AchievementPercentages struct {
		Achievements []Achievement `json:"achievements"`
	} `json:"achievementpercentages"`
}

func (s *SteamWebAPIClient) GetNews(ctx context.Context, appid string, count int, format string) (*SteamNews, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Set("appid", appid)
	params.Set("format", format)
	params.Set("count", strconv.Itoa(count))

	reqURL, err := util.BuildURL(s.url, SteamNewsURL, params)
	if err != nil {
		return nil, fmt.Errorf("get news: %w", err)
	}

	news, err := util.DoJSON[SteamNews](ctx, s.c, reqURL)
	if err != nil {
		return nil, fmt.Errorf("get news: %w", err)
	}
	return news, nil
}

func (s *SteamWebAPIClient) GetGlobalAchievementsPercentages(ctx context.Context, gameId string, format string) (*SteamAchievements, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("get global achievements: %w", err)
	}
	params := url.Values{}
	params.Set("gameid", gameId)
	params.Set("format", format)

	reqURL, err := util.BuildURL(s.url, SteamAchievementsURL, params)
	if err != nil {
		return nil, fmt.Errorf("get global achievements: %w", err)
	}

	achievements, err := util.DoJSON[SteamAchievements](ctx, s.c, reqURL)
	if err != nil {
		return nil, fmt.Errorf("get global achievements: %w", err)
	}
	return achievements, nil
}
