package steam

import (
	"encoding/json"
	"fmt"
	"strings"
)

// parseSaleHistory extracts JSON values assigned to JavaScript variables in a
// Steam Market listing HTML document. It deliberately uses a small balanced
// JSON scanner instead of a regular expression, since listings and assets can
// be very large and contain nested objects and quoted braces.
func parseSaleHistory(body []byte) (*SaleHistory, error) {
	var points [][]string
	if err := decodeEmbeddedJSON(body, "g_rgPriceHistory", &points); err != nil {
		return nil, fmt.Errorf("price history: %w", err)
	}

	history := make([]PriceHistoryEntry, 0, len(points))
	for i, point := range points {
		if len(point) < 3 {
			return nil, fmt.Errorf("price history entry %d has %d fields, want at least 3", i, len(point))
		}
		history = append(history, PriceHistoryEntry{Date: point[0], Price: point[1], Quantity: point[2]})
	}

	result := &SaleHistory{PriceHistory: history}
	// These variables are absent on some restricted or unavailable listings, so
	// they are optional even though a normal listing page includes them.
	_ = decodeEmbeddedJSON(body, "g_rgListingInfo", &result.ListingInfo)
	_ = decodeEmbeddedJSON(body, "g_rgAssets", &result.Assets)
	return result, nil
}

func decodeEmbeddedJSON(body []byte, variable string, target any) error {
	value, err := embeddedJSONValue(body, variable)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(value, target); err != nil {
		return fmt.Errorf("decode %s: %w", variable, err)
	}
	return nil
}

func embeddedJSONValue(body []byte, variable string) ([]byte, error) {
	marker := []byte(variable)
	start := strings.Index(string(body), string(marker))
	if start < 0 {
		return nil, fmt.Errorf("%s not found", variable)
	}

	i := start + len(marker)
	for i < len(body) && (body[i] == ' ' || body[i] == '\t' || body[i] == '\r' || body[i] == '\n') {
		i++
	}
	if i >= len(body) || body[i] != '=' {
		return nil, fmt.Errorf("%s has no assignment", variable)
	}
	i++
	for i < len(body) && (body[i] == ' ' || body[i] == '\t' || body[i] == '\r' || body[i] == '\n') {
		i++
	}
	if i >= len(body) || (body[i] != '[' && body[i] != '{') {
		return nil, fmt.Errorf("%s is not a JSON array or object", variable)
	}

	begin, depth := i, 0
	inString, escaped := false, false
	for ; i < len(body); i++ {
		c := body[i]
		if inString {
			if escaped {
				escaped = false
			} else if c == '\\' {
				escaped = true
			} else if c == '"' {
				inString = false
			}
			continue
		}
		switch c {
		case '"':
			inString = true
		case '[', '{':
			depth++
		case ']', '}':
			depth--
			if depth == 0 {
				return body[begin : i+1], nil
			}
		}
	}
	return nil, fmt.Errorf("%s has an unterminated JSON value", variable)
}
