package sendgrid

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestIsRetryableStatus(t *testing.T) {
	tests := []struct {
		code int
		want bool
	}{
		{http.StatusOK, false},
		{http.StatusBadRequest, false},
		{http.StatusUnauthorized, false},
		{http.StatusNotFound, false},
		{http.StatusTooManyRequests, true},
		{http.StatusInternalServerError, true},
		{http.StatusBadGateway, true},
		{http.StatusServiceUnavailable, true},
		{http.StatusGatewayTimeout, true},
	}
	for _, tt := range tests {
		if got := isRetryableStatus(tt.code); got != tt.want {
			t.Errorf("isRetryableStatus(%d) = %v, want %v", tt.code, got, tt.want)
		}
	}
}

func TestRetryDelayHonorsRetryAfter(t *testing.T) {
	got := retryDelay("2", 0, time.Second)
	if got != 2*time.Second {
		t.Errorf("retryDelay with Retry-After=2 = %v, want 2s", got)
	}
}

func TestRetryDelayExponentialBackoffWithoutRetryAfter(t *testing.T) {
	base := 100 * time.Millisecond
	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 100 * time.Millisecond},
		{1, 200 * time.Millisecond},
		{2, 400 * time.Millisecond},
	}
	for _, tt := range tests {
		if got := retryDelay("", tt.attempt, base); got != tt.want {
			t.Errorf("retryDelay(attempt=%d) = %v, want %v", tt.attempt, got, tt.want)
		}
	}
}

func TestRetryDelayCapsAtMax(t *testing.T) {
	if got := retryDelay("", 20, time.Second); got != maxRetryWait {
		t.Errorf("retryDelay at high attempt count = %v, want cap %v", got, maxRetryWait)
	}
}

func TestClientRetriesOnRateLimitThenSucceeds(t *testing.T) {
	attempts := 0
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name":"ok"}`))
	})
	client = NewClient("test-api-key", WithHost(client.host), WithMaxRetries(3), WithRetryWait(time.Millisecond))

	var out struct {
		Name string `json:"name"`
	}
	if err := client.get(context.Background(), "/v3/anything", &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
	if out.Name != "ok" {
		t.Errorf("Name = %q, want ok", out.Name)
	}
}

func TestClientDoesNotRetryOnNonRetryableStatus(t *testing.T) {
	attempts := 0
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadRequest)
	})
	client = NewClient("test-api-key", WithHost(client.host), WithMaxRetries(3), WithRetryWait(time.Millisecond))

	err := client.get(context.Background(), "/v3/anything", nil)
	if err == nil {
		t.Fatal("expected an error")
	}
	if attempts != 1 {
		t.Errorf("attempts = %d, want 1 (no retries on 400)", attempts)
	}
}

func TestClientGivesUpAfterMaxRetries(t *testing.T) {
	attempts := 0
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	client = NewClient("test-api-key", WithHost(client.host), WithMaxRetries(2), WithRetryWait(time.Millisecond))

	err := client.get(context.Background(), "/v3/anything", nil)
	if err == nil {
		t.Fatal("expected an error")
	}
	if apiErr, ok := err.(*APIError); !ok || apiErr.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("err = %v, want *APIError with status 503", err)
	}
	if attempts != 3 {
		t.Errorf("attempts = %d, want 3 (1 initial + 2 retries)", attempts)
	}
}

func TestClientSetsUserAgent(t *testing.T) {
	var gotUA string
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	})
	client = NewClient("test-api-key", WithHost(client.host), WithUserAgent("terraform-provider-sendgrid/1.2.3"))

	if err := client.get(context.Background(), "/v3/anything", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotUA != "terraform-provider-sendgrid/1.2.3" {
		t.Errorf("User-Agent = %q, want terraform-provider-sendgrid/1.2.3", gotUA)
	}
}
