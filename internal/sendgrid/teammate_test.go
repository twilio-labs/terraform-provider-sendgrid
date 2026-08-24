package sendgrid

import (
	"context"
	"net/http"
	"testing"
)

func TestInviteTeammate(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v3/teammates" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"token":"tok-123","email":"new@example.com","scopes":["user.profile.read"],"is_admin":false}`))
	})

	invite, err := client.InviteTeammate(context.Background(), "new@example.com", []string{"user.profile.read"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if invite.Token != "tok-123" || invite.Email != "new@example.com" || invite.IsAdmin {
		t.Errorf("got %+v, want token=tok-123 email=new@example.com is_admin=false", invite)
	}
}

func TestListPendingTeammateInvites(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result":[{"token":"tok-1","email":"a@example.com","scopes":[],"is_admin":true,"expiration_date":1700000000}]}`))
	})

	invites, err := client.ListPendingTeammateInvites(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(invites) != 1 || invites[0].Token != "tok-1" || invites[0].Email != "a@example.com" {
		t.Errorf("got %+v, want one invite for a@example.com", invites)
	}
}

func TestCancelTeammateInvite(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/v3/teammates/pending/tok-123" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	if err := client.CancelTeammateInvite(context.Background(), "tok-123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListTeammates(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result":[{"username":"alice","email":"alice@example.com","user_type":"teammate","is_admin":false}]}`))
	})

	teammates, err := client.ListTeammates(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(teammates) != 1 || teammates[0].Username != "alice" {
		t.Errorf("got %+v, want one teammate named alice", teammates)
	}
}

func TestGetTeammate(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/teammates/alice" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"username":"alice","email":"alice@example.com","scopes":["user.profile.read"],"is_admin":false}`))
	})

	tm, err := client.GetTeammate(context.Background(), "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tm.Username != "alice" || len(tm.Scopes) != 1 {
		t.Errorf("got %+v, want username=alice with 1 scope", tm)
	}
}

func TestUpdateTeammate(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/v3/teammates/alice" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"username":"alice","email":"alice@example.com","scopes":[],"is_admin":true}`))
	})

	tm, err := client.UpdateTeammate(context.Background(), "alice", nil, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !tm.IsAdmin {
		t.Errorf("got %+v, want is_admin=true", tm)
	}
}

func TestDeleteTeammate(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/v3/teammates/alice" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	if err := client.DeleteTeammate(context.Background(), "alice"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
