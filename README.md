# steam-market-api

A small Go client for Steam Community Market search, price-overview, and listing-page sale-history requests.

## Installation

```bash
go get github.com/olexandrf/steam-market-api
```

## Quick start

```go
client := steam.NewMarketClient(nil, steam.MarketURL)

prices, err := client.GetPrices(ctx, "730", 1, "AK-47 | Redline (Field-Tested)")
if err != nil {
    return err
}
fmt.Println(prices.LowestPrice)
```

`NewMarketClient(nil, ...)` creates an HTTP client with a five-second timeout. Pass your own `*http.Client` to configure transport, timeout, or cookies.

## API

`GetMarketList(ctx, appid, query, sortColumn, sortDir, start, count)` returns a paginated `ItemsList` from `/market/search/render`.

`GetPrices(ctx, appid, currency, marketHashName)` returns the current `PriceOverview` from `/market/priceoverview`.

`GetSaleHistory(ctx, appid, itemName, language, currency, start, count)` requests `/market/listings/...`. Steam serves this endpoint as an HTML page—not JSON. The client extracts its embedded JavaScript data:

- `PriceHistory`: date, price, and aggregated quantity for each sale point. Values remain strings because Steam localizes price formatting.
- `ListingInfo`: original JSON records for active listings, keyed by listing ID.
- `Assets`: original JSON asset records, grouped in Steam's native app/context/asset hierarchy.

```go
history, err := client.GetSaleHistory(
    ctx,
    steam.CS2ID,
    "AK-47 | Redline (Field-Tested)",
    "english",
    1, // USD
    0,
    10,
)
if err != nil {
    return err
}

for _, sale := range history.PriceHistory {
    fmt.Printf("%s: %s (%s sales)\n", sale.Date, sale.Price, sale.Quantity)
}
```

The low-level listing and asset maps use `json.RawMessage`, which lets callers decode fields that are relevant to their use case without the package losing Steam fields that may change over time.

## Errors and tests

Non-200 responses are returned as `*steam.APIError`, including the HTTP status and a bounded response-body message. Run the suite with:

```bash
go test ./...
```

## License

[MIT](LICENSE.md)
