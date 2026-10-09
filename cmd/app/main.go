package main

import (
	"context"
	"fmt"
	"log"

	"github.com/olexandrf/steam-market-web-api/steam/market"
	"github.com/olexandrf/steam-market-web-api/steam/web"
)

func logOnError(err error, msg string) {
	if err != nil {
		log.Printf("%s: %s\n", msg, err)
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var err error
	ctx := context.Background()
	webClient := web.NewSteamWebAPIClient(nil, web.SteamURL)
	steamNews, err := webClient.GetNews(ctx, web.CS2ID, 1, web.DefaultFormat)
	logOnError(err, "GetNews")
	fmt.Printf("%+v\n", steamNews)
	steamAchievements, err := webClient.GetGlobalAchievementsPercentages(ctx, web.CS2ID, web.DefaultFormat)
	logOnError(err, "GetGlobalAchievementsPercentages")
	fmt.Printf("%+v\n", steamAchievements)
	marketClient := market.NewSteamMarketClient(nil, market.SteamMarketURL)
	steamMarketResponse, err := marketClient.GetMarketList(ctx, web.CS2ID)
	logOnError(err, "SearchMarket")
	fmt.Printf("%+v\n", steamMarketResponse)
	prices, err := marketClient.GetPrices(ctx, web.CS2ID, "1", "AK-47 | Redline (Field-Tested)")
	logOnError(err, "GetPrices")
	fmt.Printf("%+v\n", prices)
}
