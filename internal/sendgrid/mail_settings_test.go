package sendgrid

import (
	"context"
	"net/http"
	"testing"
)

func TestUpdateAddressWhitelistMailSettings(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody = decodeBody(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"enabled":true,"list":["example.com"]}`))
	})

	s, err := client.UpdateAddressWhitelistMailSettings(context.Background(), AddressWhitelistMailSettings{
		Enabled: true,
		List:    []string{"example.com"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPatch || gotPath != "/v3/mail_settings/address_whitelist" {
		t.Errorf("request = %s %s, want PATCH /v3/mail_settings/address_whitelist", gotMethod, gotPath)
	}
	if gotBody["enabled"] != true {
		t.Errorf("body enabled = %v, want true", gotBody["enabled"])
	}
	if len(s.List) != 1 || s.List[0] != "example.com" {
		t.Errorf("List = %v, want [example.com]", s.List)
	}
}

func TestUpdateBouncePurgeMailSettings(t *testing.T) {
	var gotPath string
	var gotBody map[string]any
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotBody = decodeBody(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"enabled":true,"soft_bounces":7,"hard_bounces":30}`))
	})

	softBounces, hardBounces := 7, 30
	s, err := client.UpdateBouncePurgeMailSettings(context.Background(), BouncePurgeMailSettings{
		Enabled:     true,
		SoftBounces: &softBounces,
		HardBounces: &hardBounces,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/v3/mail_settings/bounce_purge" {
		t.Errorf("path = %q, want /v3/mail_settings/bounce_purge", gotPath)
	}
	if gotBody["soft_bounces"] != float64(7) {
		t.Errorf("body soft_bounces = %v, want 7", gotBody["soft_bounces"])
	}
	if s.SoftBounces == nil || *s.SoftBounces != 7 {
		t.Errorf("SoftBounces = %v, want 7", s.SoftBounces)
	}
	if s.HardBounces == nil || *s.HardBounces != 30 {
		t.Errorf("HardBounces = %v, want 30", s.HardBounces)
	}
}

func TestUpdateFooterMailSettings(t *testing.T) {
	var gotPath string
	var gotBody map[string]any
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotBody = decodeBody(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"enabled":true,"html_content":"<p>hi</p>","plain_content":"hi"}`))
	})

	s, err := client.UpdateFooterMailSettings(context.Background(), FooterMailSettings{
		Enabled:      true,
		HTMLContent:  "<p>hi</p>",
		PlainContent: "hi",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/v3/mail_settings/footer" {
		t.Errorf("path = %q, want /v3/mail_settings/footer", gotPath)
	}
	if gotBody["html_content"] != "<p>hi</p>" {
		t.Errorf("body html_content = %v, want <p>hi</p>", gotBody["html_content"])
	}
	if s.PlainContent != "hi" {
		t.Errorf("PlainContent = %q, want hi", s.PlainContent)
	}
}

func TestUpdateForwardBounceMailSettings(t *testing.T) {
	var gotPath string
	var gotBody map[string]any
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotBody = decodeBody(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"enabled":true,"email":"bounces@example.com"}`))
	})

	email := "bounces@example.com"
	s, err := client.UpdateForwardBounceMailSettings(context.Background(), ForwardBounceMailSettings{
		Enabled: true,
		Email:   &email,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/v3/mail_settings/forward_bounce" {
		t.Errorf("path = %q, want /v3/mail_settings/forward_bounce", gotPath)
	}
	if gotBody["email"] != "bounces@example.com" {
		t.Errorf("body email = %v, want bounces@example.com", gotBody["email"])
	}
	if s.Email == nil || *s.Email != "bounces@example.com" {
		t.Errorf("Email = %v, want bounces@example.com", s.Email)
	}
}

func TestUpdateForwardSpamMailSettings(t *testing.T) {
	var gotPath string
	var gotBody map[string]any
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotBody = decodeBody(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"enabled":true,"email":"spam-reports@example.com"}`))
	})

	s, err := client.UpdateForwardSpamMailSettings(context.Background(), ForwardSpamMailSettings{
		Enabled: true,
		Email:   "spam-reports@example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/v3/mail_settings/forward_spam" {
		t.Errorf("path = %q, want /v3/mail_settings/forward_spam", gotPath)
	}
	if gotBody["email"] != "spam-reports@example.com" {
		t.Errorf("body email = %v, want spam-reports@example.com", gotBody["email"])
	}
	if s.Email != "spam-reports@example.com" {
		t.Errorf("Email = %q, want spam-reports@example.com", s.Email)
	}
}

func TestUpdateTemplateMailSettings(t *testing.T) {
	var gotPath string
	var gotBody map[string]any
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotBody = decodeBody(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"enabled":true,"html_content":"<% body %>"}`))
	})

	s, err := client.UpdateTemplateMailSettings(context.Background(), TemplateMailSettings{
		Enabled:     true,
		HTMLContent: "<% body %>",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/v3/mail_settings/template" {
		t.Errorf("path = %q, want /v3/mail_settings/template", gotPath)
	}
	if gotBody["html_content"] != "<% body %>" {
		t.Errorf("body html_content = %v, want <%% body %%>", gotBody["html_content"])
	}
	if s.HTMLContent != "<% body %>" {
		t.Errorf("HTMLContent = %q, want <%% body %%>", s.HTMLContent)
	}
}
