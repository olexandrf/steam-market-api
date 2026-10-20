package steam_test

import (
	"context"
	"fmt"

	"github.com/olexandrf/steam-market-api/steam"
)

func ExampleMarketClient_GetPrices() {
	client := steam.NewMarketClient(steam.MarketURL, nil)
	price, err := client.GetPrices(context.Background(), "730", 1, "AK-47 | Redline (Field-Tested)")
	if err != nil {
		return
	}
	fmt.Println(price.LowestPrice)
}

func ExampleMarketClient_GetMarketList() {
	client := steam.NewMarketClient(steam.MarketURL, nil)
	items, err := client.GetMarketList(context.Background(), steam.CS2ID, "AK-47", "name", "asc", 0, 10)
	if err != nil {
		return
	}
	fmt.Println(items.TotalCount)
}

func ExampleMarketClient_GetSaleHistory() {
	client := steam.NewMarketClient(steam.MarketURL, nil)
	history, err := client.GetSaleHistory(context.Background(), steam.CS2ID, "AK-47 | Redline (Field-Tested)", "english", 1, 0, 10)
	if err != nil {
		return
	}
	for _, point := range history.PriceHistory {
		fmt.Printf("%s: %s (%s sales)\n", point.Date, point.Price, point.Quantity)
	}
}
