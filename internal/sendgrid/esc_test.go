package sendgrid

import (
	"context"
	"net/http"
	"testing"
)

// TestEscPreventsPathAndQueryInjection verifies that an identifier
// containing "/" or "?" can't redirect a request onto a different path or
// inject query parameters, by round-tripping it through a real client call.
func TestEscPreventsPathAndQueryInjection(t *testing.T) {
	var gotPath, gotQuery string
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
	})

	maliciousID := "abc?evil=1&other=/v3/other-endpoint"
	if err := client.get(context.Background(), "/v3/api_keys/"+esc(maliciousID), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotQuery != "" {
		t.Errorf("RawQuery = %q, want empty (the '?' should have been escaped into the path, not treated as a delimiter)", gotQuery)
	}
	if gotPath != "/v3/api_keys/"+maliciousID {
		t.Errorf("Path = %q, want the ID preserved verbatim as a single path segment", gotPath)
	}

	slashID := "foo/../bar"
	gotPath = ""
	if err := client.get(context.Background(), "/v3/api_keys/"+esc(slashID), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/v3/api_keys/"+slashID {
		t.Errorf("Path = %q, want the ID's slashes preserved as literal characters in one segment, not a path traversal", gotPath)
	}
}
