package main

import (
	"context"
	"fmt"
	"log"

	"github.com/olexandrf/steam-market-api/steam"
)

func main() {
	client := steam.NewMarketClient(steam.MarketURL, nil)

	// itemName must be the Steam Market hash name, not an asset ID or a listing ID.
	history, err := client.GetSaleHistory(
		context.Background(),
		steam.CS2ID,
		"AK-47 | Redline (Field-Tested)",
		"english",
		1, // USD
		0,
		10,
	)
	if err != nil {
		log.Fatal(err)
	}

	for _, sale := range history.PriceHistory {
		fmt.Printf("%s: %s (%s sales)\n", sale.Date, sale.Price, sale.Quantity)
	}

	// ListingInfo and Assets retain Steam's original JSON. For example, use a
	// listing ID as the map key, then unmarshal its json.RawMessage into a type
	// containing the Steam fields your application needs.
	fmt.Printf("active listings: %d\n", len(history.ListingInfo))
}
