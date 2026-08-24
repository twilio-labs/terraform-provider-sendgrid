package sendgrid

import (
	"context"
	"net/http"
	"testing"
)

func TestDeleteListQueryParameter(t *testing.T) {
	var gotQuery string
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusNoContent)
	})

	if err := client.DeleteList(context.Background(), "abc123", true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotQuery != "delete_contacts=true" {
		t.Errorf("RawQuery = %q, want %q", gotQuery, "delete_contacts=true")
	}

	if err := client.DeleteList(context.Background(), "abc123", false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotQuery != "delete_contacts=false" {
		t.Errorf("RawQuery = %q, want %q", gotQuery, "delete_contacts=false")
	}
}
