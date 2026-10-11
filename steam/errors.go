package steam

import (
	"errors"
	"fmt"
)

var ErrClientNil = errors.New("client is nil")

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("http %d", e.StatusCode)
	}
	return fmt.Sprintf("http %d: %s", e.StatusCode, e.Message)
}
