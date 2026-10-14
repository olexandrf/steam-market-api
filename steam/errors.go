package steam

import (
	"errors"
	"fmt"
	"time"
)

var ErrClientNil = errors.New("client is nil")

type APIError struct {
	StatusCode int
	Message    string
	RetryAfter time.Duration
}

func (e *APIError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("http %d", e.StatusCode)
	}
	return fmt.Sprintf("http %d: %s", e.StatusCode, e.Message)
}

type URLError struct {
	Err error
}

func (e *URLError) Error() string {
	return "request error: " + e.Err.Error()
}
