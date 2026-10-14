package steam

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMarketClient_GetMarketList(t *testing.T) {
	body := `{}`
	item := ""
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	}))

	defer srv.Close()
	client := NewMarketClient(srv.Client(), srv.URL)
	items, err := client.GetMarketList(context.Background(), CS2ID, item)
	if err != nil {
		t.Errorf("GetMarketList() error = %v", err)
	}

	if items.Success != true {
		t.Errorf("GetMarketList() error = %v", err)
	}
	if items.TotalCount != 2 {
		t.Errorf("GetMarketList() error = %v", err)
	}
	if len(items.Results) != 2 {
		t.Errorf("GetMarketList() error = %v", err)
	}
}

func TestMarketClient_GetMarketList_Errors(t *testing.T) {

	tests := []struct {
		name           string
		status         int
		body           string
		isAPIError     bool // 0 if error returned not from API
		expectedErrMsg string
	}{
		{
			name:       "status 429",
			status:     http.StatusTooManyRequests,
			body:       `{'success': false, 'message': 'API rate limit exceeded'}`,
			isAPIError: true,
		},
		{
			name:       "status 400",
			status:     http.StatusBadRequest,
			body:       `{'success': false, 'message': 'Invalid request'}`,
			isAPIError: true,
		},
		{
			name:       "status 500",
			status:     http.StatusInternalServerError,
			body:       `{'success': false, 'message': 'Internal Server Error'}`,
			isAPIError: true,
		},
		{
			name:       "status 503",
			status:     http.StatusServiceUnavailable,
			body:       `{'success': false, 'message': 'Service unavailable'}`,
			isAPIError: true,
		},
		{
			name:       "status 500 and message is too big",
			status:     http.StatusInternalServerError,
			body:       `{'success': false, 'message': 'Very long message'}`,
			isAPIError: true,
		},
		{
			name:       "invalid json",
			status:     http.StatusOK,
			body:       `{"success":true,`,
			isAPIError: false,
		},
		{
			name:       "empty body",
			status:     http.StatusOK,
			body:       ``,
			isAPIError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(test.status)
				w.Write([]byte(test.body))
			}))
			defer srv.Close()
			_, err := NewMarketClient(srv.Client(), srv.URL).GetMarketList(context.Background(), CS2ID, "testItem")
			if test.isAPIError == true {
				err, ok := errors.AsType[*APIError](err)
				if !ok {
					t.Fatal("expected APIError")
				}
				if err.StatusCode != test.status {
					t.Fatalf("expected APIError status %d, got %d", test.status, err.StatusCode)
				}
			} else {
				if err == nil {
					t.Fatal("expected error")
				}

			}
		})
	}
}
