package sendgrid

import (
	"context"
	"net/http"
	"testing"
)

func TestGetSubuserByUsernameExactMatch(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Includes a username that is a prefix of the target, to verify the
		// client performs an exact match rather than trusting the API's
		// query-param filtering alone.
		_, _ = w.Write([]byte(`[
			{"id": 1, "username": "alice", "email": "alice@example.com", "disabled": false},
			{"id": 2, "username": "alice-admin", "email": "alice-admin@example.com", "disabled": false}
		]`))
	})

	sub, err := client.GetSubuserByUsername(context.Background(), "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub == nil {
		t.Fatal("expected a matching subuser, got nil")
	}
	if sub.ID != 1 || sub.Username != "alice" || sub.Email != "alice@example.com" {
		t.Errorf("got %+v, want ID=1 Username=alice Email=alice@example.com", sub)
	}
}

func TestGetSubuserByUsernameNotFound(t *testing.T) {
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	})

	sub, err := client.GetSubuserByUsername(context.Background(), "nobody")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub != nil {
		t.Errorf("expected nil for no match, got %+v", sub)
	}
}
