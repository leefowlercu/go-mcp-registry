package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestNewRequest(t *testing.T) {
	c := NewClient(nil)

	tests := []struct {
		name        string
		baseURL     string
		method      string
		urlStr      string
		body        any
		wantErr     bool
		wantErrMsg  string
		checkHeader bool
	}{
		{
			name:    "valid request without body",
			baseURL: "https://api.example.com/",
			method:  "GET",
			urlStr:  "v0.1/servers",
			body:    nil,
			wantErr: false,
		},
		{
			name:    "valid request with body",
			baseURL: "https://api.example.com/",
			method:  "POST",
			urlStr:  "v0.1/servers",
			body:    map[string]string{"name": "test"},
			wantErr: false,
		},
		{
			name:       "baseURL without trailing slash",
			baseURL:    "https://api.example.com",
			method:     "GET",
			urlStr:     "v0.1/servers",
			body:       nil,
			wantErr:    true,
			wantErrMsg: "BaseURL must have a trailing slash",
		},
		{
			name:       "invalid URL path",
			baseURL:    "https://api.example.com/",
			method:     "GET",
			urlStr:     "://invalid",
			body:       nil,
			wantErr:    true,
			wantErrMsg: "parse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the base URL
			baseURL, _ := url.Parse(tt.baseURL)
			c.BaseURL = baseURL

			req, err := c.NewRequest(tt.method, tt.urlStr, tt.body)

			if tt.wantErr {
				if err == nil {
					t.Error("NewRequest() expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("NewRequest() error = %q, want to contain %q", err.Error(), tt.wantErrMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("NewRequest() unexpected error: %v", err)
				return
			}

			if req == nil {
				t.Fatal("NewRequest() returned nil request")
			}

			if req.Method != tt.method {
				t.Errorf("NewRequest() method = %q, want %q", req.Method, tt.method)
			}

			if tt.body != nil {
				if req.Header.Get("Content-Type") != mediaTypeJSON {
					t.Errorf("NewRequest() Content-Type = %q, want %q", req.Header.Get("Content-Type"), mediaTypeJSON)
				}
			}

			if req.Header.Get("Accept") != mediaTypeJSON {
				t.Errorf("NewRequest() Accept = %q, want %q", req.Header.Get("Accept"), mediaTypeJSON)
			}

			if req.Header.Get("User-Agent") == "" {
				t.Error("NewRequest() User-Agent header not set")
			}
		})
	}
}

func TestNewRequest_BadJSON(t *testing.T) {
	c := NewClient(nil)

	// Create a type that can't be marshaled to JSON
	type InvalidJSON struct {
		BadField chan int // channels can't be marshaled to JSON
	}

	_, err := c.NewRequest("POST", "v0.1/servers", &InvalidJSON{BadField: make(chan int)})
	if err == nil {
		t.Error("NewRequest() expected JSON encoding error, got nil")
	}
}

func TestParseRate(t *testing.T) {
	tests := []struct {
		name     string
		headers  http.Header
		wantRate Rate
	}{
		{
			name:     "no rate limit headers",
			headers:  http.Header{},
			wantRate: Rate{},
		},
		{
			name: "all rate limit headers present",
			headers: http.Header{
				"X-Ratelimit-Limit":     []string{"100"},
				"X-Ratelimit-Remaining": []string{"50"},
				"X-Ratelimit-Reset":     []string{"2024-01-01T12:00:00Z"},
			},
			wantRate: Rate{
				Limit:     100,
				Remaining: 50,
				Reset:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "only limit header",
			headers: http.Header{
				"X-Ratelimit-Limit": []string{"100"},
			},
			wantRate: Rate{
				Limit:     100,
				Remaining: 0,
				Reset:     time.Time{},
			},
		},
		{
			name: "invalid reset timestamp",
			headers: http.Header{
				"X-Ratelimit-Limit":     []string{"100"},
				"X-Ratelimit-Remaining": []string{"50"},
				"X-Ratelimit-Reset":     []string{"invalid"},
			},
			wantRate: Rate{
				Limit:     100,
				Remaining: 50,
				Reset:     time.Time{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				Header: tt.headers,
			}

			rate := parseRate(resp)

			if rate.Limit != tt.wantRate.Limit {
				t.Errorf("parseRate() Limit = %d, want %d", rate.Limit, tt.wantRate.Limit)
			}
			if rate.Remaining != tt.wantRate.Remaining {
				t.Errorf("parseRate() Remaining = %d, want %d", rate.Remaining, tt.wantRate.Remaining)
			}
			if !rate.Reset.Equal(tt.wantRate.Reset) {
				t.Errorf("parseRate() Reset = %v, want %v", rate.Reset, tt.wantRate.Reset)
			}
		})
	}
}

func TestDo(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		statusCode  int
		responseBody string
		wantErr     bool
		wantErrMsg  string
	}{
		{
			name:         "successful request",
			ctx:          context.Background(),
			statusCode:   200,
			responseBody: `{"name": "test"}`,
			wantErr:      false,
		},
		{
			name:         "error response",
			ctx:          context.Background(),
			statusCode:   404,
			responseBody: `{"message": "Not found"}`,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				fmt.Fprint(w, tt.responseBody)
			}))
			defer server.Close()

			client := NewClient(nil)
			client.BaseURL, _ = url.Parse(server.URL + "/")

			req, _ := client.NewRequest("GET", "test", nil)

			var result map[string]string
			_, err := client.Do(tt.ctx, req, &result)

			if tt.wantErr {
				if err == nil {
					t.Error("Do() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Do() unexpected error: %v", err)
			}
		})
	}
}

func TestDo_NilContext(t *testing.T) {
	client := NewClient(nil)
	req, _ := client.NewRequest("GET", "test", nil)

	_, err := client.Do(nil, req, nil)
	if err == nil {
		t.Error("Do() with nil context expected error, got nil")
	}
	if !strings.Contains(err.Error(), "context must be non-nil") {
		t.Errorf("Do() error = %q, want to contain %q", err.Error(), "context must be non-nil")
	}
}

func TestDo_CancelledContext(t *testing.T) {
	// Create a server that delays the response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(200)
	}))
	defer server.Close()

	client := NewClient(nil)
	client.BaseURL, _ = url.Parse(server.URL + "/")
	req, _ := client.NewRequest("GET", "test", nil)

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.Do(ctx, req, nil)
	if err == nil {
		t.Error("Do() with cancelled context expected error, got nil")
	}
}

func TestDo_IOWriter(t *testing.T) {
	responseBody := "raw response body"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprint(w, responseBody)
	}))
	defer server.Close()

	client := NewClient(nil)
	client.BaseURL, _ = url.Parse(server.URL + "/")
	req, _ := client.NewRequest("GET", "test", nil)

	var buf bytes.Buffer
	_, err := client.Do(context.Background(), req, &buf)
	if err != nil {
		t.Errorf("Do() with io.Writer unexpected error: %v", err)
	}

	if buf.String() != responseBody {
		t.Errorf("Do() wrote %q to io.Writer, want %q", buf.String(), responseBody)
	}
}

func TestDo_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer server.Close()

	client := NewClient(nil)
	client.BaseURL, _ = url.Parse(server.URL + "/")
	req, _ := client.NewRequest("GET", "test", nil)

	var result map[string]string
	_, err := client.Do(context.Background(), req, &result)
	if err != nil {
		t.Errorf("Do() with empty response unexpected error: %v", err)
	}
}

func TestDo_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprint(w, "not valid json")
	}))
	defer server.Close()

	client := NewClient(nil)
	client.BaseURL, _ = url.Parse(server.URL + "/")
	req, _ := client.NewRequest("GET", "test", nil)

	var result map[string]string
	_, err := client.Do(context.Background(), req, &result)
	if err == nil {
		t.Error("Do() with invalid JSON expected error, got nil")
	}
	if _, ok := err.(*json.SyntaxError); !ok {
		t.Errorf("Do() error type = %T, want *json.SyntaxError", err)
	}
}

func TestAddOptions(t *testing.T) {
	type options struct {
		Limit  int    `url:"limit,omitempty"`
		Cursor string `url:"cursor,omitempty"`
		Search string `url:"search,omitempty"`
	}

	tests := []struct {
		name     string
		baseURL  string
		opts     any
		wantURL  string
		wantErr  bool
	}{
		{
			name:    "no options",
			baseURL: "v0.1/servers",
			opts:    nil,
			wantURL: "v0.1/servers",
			wantErr: false,
		},
		{
			name:    "with options",
			baseURL: "v0.1/servers",
			opts: &options{
				Limit:  10,
				Cursor: "abc123",
			},
			wantURL: "v0.1/servers?cursor=abc123&limit=10",
			wantErr: false,
		},
		{
			name:    "existing query parameters",
			baseURL: "v0.1/servers?existing=param",
			opts: &options{
				Limit: 10,
			},
			wantURL: "v0.1/servers?existing=param&limit=10",
			wantErr: false,
		},
		{
			name:    "all fields empty",
			baseURL: "v0.1/servers",
			opts:    &options{},
			wantURL: "v0.1/servers",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := addOptions(tt.baseURL, tt.opts)

			if tt.wantErr {
				if err == nil {
					t.Error("addOptions() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("addOptions() unexpected error: %v", err)
				return
			}

			if got != tt.wantURL {
				t.Errorf("addOptions() = %q, want %q", got, tt.wantURL)
			}
		})
	}
}

func TestAddOptions_InvalidURL(t *testing.T) {
	opts := &ServerListOptions{
		Search: "test",
	}

	_, err := addOptions("://invalid", opts)
	if err == nil {
		t.Error("addOptions() with invalid URL expected error, got nil")
	}
}

func TestNewResponse(t *testing.T) {
	resetTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	httpResp := &http.Response{
		Header: http.Header{
			"X-Ratelimit-Limit":     []string{"100"},
			"X-Ratelimit-Remaining": []string{"50"},
			"X-Ratelimit-Reset":     []string{resetTime.Format(time.RFC3339)},
		},
	}

	resp := newResponse(httpResp)

	if resp.Response != httpResp {
		t.Error("newResponse() did not set Response field correctly")
	}

	if resp.Rate.Limit != 100 {
		t.Errorf("newResponse() Rate.Limit = %d, want 100", resp.Rate.Limit)
	}

	if resp.Rate.Remaining != 50 {
		t.Errorf("newResponse() Rate.Remaining = %d, want 50", resp.Rate.Remaining)
	}

	if !resp.Rate.Reset.Equal(resetTime) {
		t.Errorf("newResponse() Rate.Reset = %v, want %v", resp.Rate.Reset, resetTime)
	}
}
