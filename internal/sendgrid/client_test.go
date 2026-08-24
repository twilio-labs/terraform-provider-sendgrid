package sendgrid

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testServer(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return NewClient("test-api-key", WithHost(srv.URL)), srv
}

func TestClientAuthorizationHeader(t *testing.T) {
	var gotAuth string
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	})

	if err := client.get(context.Background(), "/v3/anything", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := "Bearer test-api-key"; gotAuth != want {
		t.Errorf("Authorization header = %q, want %q", gotAuth, want)
	}
}

func TestClientOnBehalfOfHeader(t *testing.T) {
	var gotHeader string
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("On-Behalf-Of")
		w.WriteHeader(http.StatusOK)
	})
	client = NewClient("test-api-key", WithHost(client.host), WithOnBehalfOf("subuser1"))

	if err := client.get(context.Background(), "/v3/anything", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotHeader != "subuser1" {
		t.Errorf("On-Behalf-Of header = %q, want %q", gotHeader, "subuser1")
	}
}

func TestClientDecodesSuccessResponse(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name":"hello"}`))
	})

	var out struct {
		Name string `json:"name"`
	}
	if err := client.get(context.Background(), "/v3/anything", &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Name != "hello" {
		t.Errorf("Name = %q, want %q", out.Name, "hello")
	}
}

func TestClientReturnsAPIErrorOnNon2xx(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"errors":[{"message":"bad request","field":"name"}]}`))
	})

	err := client.get(context.Background(), "/v3/anything", nil)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusBadRequest)
	}
	if len(apiErr.Errors) != 1 || apiErr.Errors[0].Message != "bad request" {
		t.Errorf("Errors = %+v, want a single 'bad request' error", apiErr.Errors)
	}
}

func TestIsNotFound(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"message":"not found"}]}`))
	})

	err := client.get(context.Background(), "/v3/anything", nil)
	if !IsNotFound(err) {
		t.Errorf("IsNotFound(%v) = false, want true", err)
	}

	if IsNotFound(nil) {
		t.Error("IsNotFound(nil) = true, want false")
	}
	if IsNotFound(context.DeadlineExceeded) {
		t.Error("IsNotFound on a non-APIError = true, want false")
	}

	wrapped := fmt.Errorf("while doing something: %w", err)
	if !IsNotFound(wrapped) {
		t.Error("IsNotFound on a wrapped *APIError = false, want true")
	}
}

func TestClientQueryParameters(t *testing.T) {
	var gotQuery string
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
	})

	err := client.get(context.Background(), "/v3/anything", nil, Query("username", "alice"), Query("limit", "10"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotQuery != "limit=10&username=alice" {
		t.Errorf("RawQuery = %q, want %q", gotQuery, "limit=10&username=alice")
	}
}

func TestClientSendsJSONBodyOnlyWhenPresent(t *testing.T) {
	var gotContentType string
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusNoContent)
	})

	if err := client.delete(context.Background(), "/v3/anything"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotContentType != "" {
		t.Errorf("Content-Type = %q on a bodyless request, want empty", gotContentType)
	}
}
