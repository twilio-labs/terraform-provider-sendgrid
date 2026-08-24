package sendgrid

import (
	"context"
	"net/http"
	"testing"
)

func TestUpdateClickTrackingSettings(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody = decodeBody(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"enabled":true,"enable_text":false}`))
	})

	s, err := client.UpdateClickTrackingSettings(context.Background(), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPatch || gotPath != "/v3/tracking_settings/click" {
		t.Errorf("request = %s %s, want PATCH /v3/tracking_settings/click", gotMethod, gotPath)
	}
	if gotBody["enabled"] != true {
		t.Errorf("body enabled = %v, want true", gotBody["enabled"])
	}
	if !s.Enabled {
		t.Errorf("Enabled = %v, want true", s.Enabled)
	}
}

func TestUpdateOpenTrackingSettings(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody = decodeBody(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"enabled":false}`))
	})

	s, err := client.UpdateOpenTrackingSettings(context.Background(), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPatch || gotPath != "/v3/tracking_settings/open" {
		t.Errorf("request = %s %s, want PATCH /v3/tracking_settings/open", gotMethod, gotPath)
	}
	if gotBody["enabled"] != false {
		t.Errorf("body enabled = %v, want false", gotBody["enabled"])
	}
	if s.Enabled {
		t.Errorf("Enabled = %v, want false", s.Enabled)
	}
}

func TestUpdateSubscriptionTrackingSettings(t *testing.T) {
	var gotPath string
	var gotBody map[string]any
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotBody = decodeBody(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"enabled":true,"url":"https://example.com/unsub"}`))
	})

	s, err := client.UpdateSubscriptionTrackingSettings(context.Background(), SubscriptionTrackingSettings{
		Enabled: true,
		URL:     "https://example.com/unsub",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/v3/tracking_settings/subscription" {
		t.Errorf("path = %q, want /v3/tracking_settings/subscription", gotPath)
	}
	if gotBody["enabled"] != true || gotBody["url"] != "https://example.com/unsub" {
		t.Errorf("body = %v, want enabled=true url=https://example.com/unsub", gotBody)
	}
	if s.URL != "https://example.com/unsub" {
		t.Errorf("URL = %q, want https://example.com/unsub", s.URL)
	}
}

func TestUpdateGoogleAnalyticsSettings(t *testing.T) {
	var gotPath string
	var gotBody map[string]any
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotBody = decodeBody(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"enabled":true,"utm_source":"sendgrid.com"}`))
	})

	s, err := client.UpdateGoogleAnalyticsSettings(context.Background(), GoogleAnalyticsSettings{
		Enabled:   true,
		UTMSource: "sendgrid.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/v3/tracking_settings/google_analytics" {
		t.Errorf("path = %q, want /v3/tracking_settings/google_analytics", gotPath)
	}
	if gotBody["utm_source"] != "sendgrid.com" {
		t.Errorf("body utm_source = %v, want sendgrid.com", gotBody["utm_source"])
	}
	if s.UTMSource != "sendgrid.com" {
		t.Errorf("UTMSource = %q, want sendgrid.com", s.UTMSource)
	}
}
