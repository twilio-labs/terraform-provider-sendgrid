package sendgrid

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
)

func TestCreatePartnerAccountRequestAndResponse(t *testing.T) {
	var gotBody createPartnerAccountRequest
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v3/partners/accounts" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"account_id":"sg2a2bcd3ef4ab5c67d8efab91c01de2fa"}`))
	})

	profile := &PartnerAccountProfile{FirstName: "Sender", Email: "mail@example.com"}
	offerings := []PartnerOffering{{Name: "org.ei.free.v1", Type: "package", Quantity: 1}}

	accountID, err := client.CreatePartnerAccount(context.Background(), profile, offerings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if accountID != "sg2a2bcd3ef4ab5c67d8efab91c01de2fa" {
		t.Errorf("accountID = %q, want %q", accountID, "sg2a2bcd3ef4ab5c67d8efab91c01de2fa")
	}
	if gotBody.Profile == nil || gotBody.Profile.Email != "mail@example.com" {
		t.Errorf("request profile = %+v, want email mail@example.com", gotBody.Profile)
	}
	if len(gotBody.Offerings) != 1 || gotBody.Offerings[0].Name != "org.ei.free.v1" {
		t.Errorf("request offerings = %+v", gotBody.Offerings)
	}
}

func TestListPartnerAccountsPagination(t *testing.T) {
	calls := 0
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		switch calls {
		case 1:
			if got := r.URL.Query().Get("offset"); got != "" {
				t.Errorf("first call offset = %q, want empty", got)
			}
			_, _ = w.Write([]byte(`{"accounts":[` + accountsJSON(100, 0) + `],"pages":{"last":"acct-99"}}`))
		case 2:
			if got := r.URL.Query().Get("offset"); got != "acct-99" {
				t.Errorf("second call offset = %q, want acct-99", got)
			}
			_, _ = w.Write([]byte(`{"accounts":[{"id":"acct-100","created_at":"2024-01-01T00:00:00Z"}],"pages":{"last":"acct-100"}}`))
		default:
			t.Fatalf("unexpected third call")
		}
	})

	accounts, err := client.ListPartnerAccounts(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(accounts) != 101 {
		t.Errorf("len(accounts) = %d, want 101", len(accounts))
	}
	if calls != 2 {
		t.Errorf("calls = %d, want 2", calls)
	}
}

func accountsJSON(n, startAt int) string {
	out := ""
	for i := startAt; i < n; i++ {
		if i > startAt {
			out += ","
		}
		out += `{"id":"acct-` + strconv.Itoa(i) + `","created_at":"2024-01-01T00:00:00Z"}`
	}
	return out
}

func TestSetPartnerAccountState(t *testing.T) {
	var gotPath, gotMethod string
	var gotBody partnerAccountStateResponse
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
	})

	if err := client.SetPartnerAccountState(context.Background(), "acct-1", "deactivated"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPut {
		t.Errorf("method = %q, want PUT", gotMethod)
	}
	if gotPath != "/v3/partners/accounts/acct-1/state" {
		t.Errorf("path = %q", gotPath)
	}
	if gotBody.State != "deactivated" {
		t.Errorf("state = %q, want deactivated", gotBody.State)
	}
}

func TestDeletePartnerAccount(t *testing.T) {
	var gotPath, gotMethod string
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.WriteHeader(http.StatusNoContent)
	})

	if err := client.DeletePartnerAccount(context.Background(), "acct-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("method = %q, want DELETE", gotMethod)
	}
	if want := "/v3/partners/accounts/acct-1"; gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
}
