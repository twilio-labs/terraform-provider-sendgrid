package sendgrid

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestScheduleSingleSend(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"send_at":"2026-03-01T09:00:00Z","status":"scheduled"}`))
	})

	status, err := client.ScheduleSingleSend(context.Background(), "id1", "2026-03-01T09:00:00Z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPut {
		t.Errorf("method = %q, want PUT", gotMethod)
	}
	if gotPath != "/v3/marketing/singlesends/id1/schedule" {
		t.Errorf("path = %q, want /v3/marketing/singlesends/id1/schedule", gotPath)
	}
	if gotBody["send_at"] != "2026-03-01T09:00:00Z" {
		t.Errorf("body send_at = %v, want 2026-03-01T09:00:00Z", gotBody["send_at"])
	}
	if status != "scheduled" {
		t.Errorf("status = %q, want scheduled", status)
	}
}

func TestScheduleSingleSendNow(t *testing.T) {
	var gotBody map[string]any
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"send_at":"now","status":"triggered"}`))
	})

	status, err := client.ScheduleSingleSend(context.Background(), "id1", "now")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["send_at"] != "now" {
		t.Errorf("body send_at = %v, want now (verbatim, unescaped)", gotBody["send_at"])
	}
	if status != "triggered" {
		t.Errorf("status = %q, want triggered", status)
	}
}

func TestUnscheduleSingleSend(t *testing.T) {
	var gotMethod, gotPath string
	client, _ := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})

	if err := client.UnscheduleSingleSend(context.Background(), "id1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("method = %q, want DELETE", gotMethod)
	}
	if gotPath != "/v3/marketing/singlesends/id1/schedule" {
		t.Errorf("path = %q, want /v3/marketing/singlesends/id1/schedule", gotPath)
	}
}
