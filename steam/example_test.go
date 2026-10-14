package steam_test

import (
	"context"
	"fmt"
	"log"

	"github.com/olexandrf/steam-market-web-api/steam"
)

func ExampleMarketClient_GetPrices() {
	SteamMarketClient := steam.NewMarketClient(nil, steam.MarketURL)
	steamPriceOverview, err := SteamMarketClient.GetPrices(context.Background(), steam.CS2ID, "1", "AK-47 | Redline (Field-Tested)")
	if err != nil {
		log.Fatalf(err.Error())
	}
	fmt.Printf("%+v\n", steamPriceOverview)
}

func ExampleMarketClient_GetMarketList() {
	SteamMarketClient := steam.NewMarketClient(nil, steam.MarketURL)
	steamMarketResponse, err := SteamMarketClient.GetMarketList(context.Background(), steam.CS2ID)
	if err != nil {
		log.Fatalf(err.Error())
	}
	fmt.Printf("%+v\n", steamMarketResponse)
}
