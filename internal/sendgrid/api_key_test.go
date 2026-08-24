package sendgrid

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func decodeBody(t *testing.T, r *http.Request) map[string]any {
	t.Helper()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("reading request body: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("decoding request body %q: %v", b, err)
	}
	return out
}

func TestCreateAPIKeyOmitsScopesWhenEmpty(t *testing.T) {
	var gotBody map[string]any
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody = decodeBody(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"api_key_id":"id1","name":"key"}`))
	})

	if _, err := client.CreateAPIKey(context.Background(), "key", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := gotBody["scopes"]; ok {
		t.Errorf("body = %v, expected no \"scopes\" key when scopes is empty", gotBody)
	}
}

func TestCreateAPIKeyIncludesScopesWhenSet(t *testing.T) {
	var gotBody map[string]any
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody = decodeBody(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"api_key_id":"id1","name":"key","scopes":["mail.send"]}`))
	})

	if _, err := client.CreateAPIKey(context.Background(), "key", []string{"mail.send"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	scopes, ok := gotBody["scopes"].([]any)
	if !ok || len(scopes) != 1 || scopes[0] != "mail.send" {
		t.Errorf("body scopes = %v, want [mail.send]", gotBody["scopes"])
	}
}

func TestUpdateAPIKeyRequiresScopes(t *testing.T) {
	var gotBody map[string]any
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody = decodeBody(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"api_key_id":"id1","name":"key","scopes":["mail.send"]}`))
	})

	if _, err := client.UpdateAPIKey(context.Background(), "id1", "key", []string{"mail.send"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	scopes, ok := gotBody["scopes"].([]any)
	if !ok || len(scopes) != 1 || scopes[0] != "mail.send" {
		t.Errorf("body scopes = %v, want [mail.send]", gotBody["scopes"])
	}
}

func TestUpdateAPIKeyNameOmitsScopes(t *testing.T) {
	var gotBody map[string]any
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody = decodeBody(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"api_key_id":"id1","name":"renamed"}`))
	})

	if _, err := client.UpdateAPIKeyName(context.Background(), "id1", "renamed"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := gotBody["scopes"]; ok {
		t.Errorf("body = %v, expected no \"scopes\" key from the rename-only endpoint", gotBody)
	}
	if gotBody["name"] != "renamed" {
		t.Errorf("body name = %v, want renamed", gotBody["name"])
	}
}
