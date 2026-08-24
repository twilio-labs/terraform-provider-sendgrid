package sendgrid

import (
	"context"
	"net/http"
	"testing"
)

func TestStartIPWarmup(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody = decodeBody(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"ip":"1.2.3.4","start_date":1409616000}]`))
	})

	w, err := client.StartIPWarmup(context.Background(), "1.2.3.4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/v3/ips/warmup" {
		t.Errorf("request = %s %s, want POST /v3/ips/warmup", gotMethod, gotPath)
	}
	if gotBody["ip"] != "1.2.3.4" {
		t.Errorf("body ip = %v, want 1.2.3.4", gotBody["ip"])
	}
	if w == nil || w.StartDate != 1409616000 {
		t.Errorf("StartIPWarmup = %v, want start_date 1409616000", w)
	}
}

func TestGetIPWarmupEmptyMeansNotInWarmup(t *testing.T) {
	var gotPath string
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	})

	w, err := client.GetIPWarmup(context.Background(), "1.2.3.4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w != nil {
		t.Errorf("GetIPWarmup = %v, want nil for empty array response", w)
	}
	if gotPath != "/v3/ips/warmup/1.2.3.4" {
		t.Errorf("path = %q, want /v3/ips/warmup/1.2.3.4", gotPath)
	}
}

func TestStopIPWarmup(t *testing.T) {
	var gotMethod, gotPath string
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})

	if err := client.StopIPWarmup(context.Background(), "1.2.3.4"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodDelete || gotPath != "/v3/ips/warmup/1.2.3.4" {
		t.Errorf("request = %s %s, want DELETE /v3/ips/warmup/1.2.3.4", gotMethod, gotPath)
	}
}

func TestFirstMatchingIPWarmup(t *testing.T) {
	list := []IPWarmup{{IP: "1.1.1.1", StartDate: 1}, {IP: "2.2.2.2", StartDate: 2}}
	if got := firstMatchingIPWarmup(list, "2.2.2.2"); got == nil || got.StartDate != 2 {
		t.Errorf("firstMatchingIPWarmup exact match = %v, want start_date 2", got)
	}
	if got := firstMatchingIPWarmup(list, "3.3.3.3"); got != nil {
		t.Errorf("firstMatchingIPWarmup no match in multi-item list = %v, want nil", got)
	}
	single := []IPWarmup{{IP: "9.9.9.9", StartDate: 9}}
	if got := firstMatchingIPWarmup(single, "9.9.9.9"); got == nil || got.StartDate != 9 {
		t.Errorf("firstMatchingIPWarmup single-item fallback = %v, want start_date 9", got)
	}
	if got := firstMatchingIPWarmup(nil, "1.1.1.1"); got != nil {
		t.Errorf("firstMatchingIPWarmup empty list = %v, want nil", got)
	}
}
