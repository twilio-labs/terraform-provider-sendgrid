package sendgrid

import (
	"context"
	"net/http"
	"testing"
)

func TestAddIPAddress(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody = decodeBody(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"ip":"1.2.3.4","is_auto_warmup":true,"is_parent_assigned":false,"subusers":["123"]}`))
	})

	ip, err := client.AddIPAddress(context.Background(), true, false, []string{"123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/v3/send_ips/ips" {
		t.Errorf("request = %s %s, want POST /v3/send_ips/ips", gotMethod, gotPath)
	}
	if gotBody["is_auto_warmup"] != true {
		t.Errorf("body is_auto_warmup = %v, want true", gotBody["is_auto_warmup"])
	}
	if ip.IP != "1.2.3.4" {
		t.Errorf("IP = %q, want 1.2.3.4", ip.IP)
	}
}

func TestAddIPAddressOmitsEmptySubusers(t *testing.T) {
	var gotBody map[string]any
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody = decodeBody(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"ip":"1.2.3.4"}`))
	})

	if _, err := client.AddIPAddress(context.Background(), false, false, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := gotBody["subusers"]; ok {
		t.Errorf("body contains subusers = %v, want omitted", gotBody["subusers"])
	}
}

func TestGetIPAddress(t *testing.T) {
	var gotPath string
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"ip":"1.2.3.4",
			"is_parent_assigned":true,
			"is_auto_warmup":false,
			"pools":[{"id":"1","name":"marketing"}],
			"added_at":1409616000,
			"updated_at":null,
			"is_enabled":null,
			"is_leased":true
		}`))
	})

	d, err := client.GetIPAddress(context.Background(), "1.2.3.4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/v3/send_ips/ips/1.2.3.4" {
		t.Errorf("path = %q, want /v3/send_ips/ips/1.2.3.4", gotPath)
	}
	if len(d.Pools) != 1 || d.Pools[0].Name != "marketing" {
		t.Errorf("Pools = %v, want [{marketing}]", d.Pools)
	}
	if d.UpdatedAt != nil {
		t.Errorf("UpdatedAt = %v, want nil", d.UpdatedAt)
	}
	if !d.IsLeased {
		t.Errorf("IsLeased = false, want true")
	}
}

func TestGetIPAddressEscapesPath(t *testing.T) {
	var gotEscapedPath string
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotEscapedPath = r.URL.EscapedPath()
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"message":"not found"}]}`))
	})

	_, err := client.GetIPAddress(context.Background(), "not-a-real-ip/../etc")
	if !IsNotFound(err) {
		t.Fatalf("expected IsNotFound, got: %v", err)
	}
	if gotEscapedPath != "/v3/send_ips/ips/not-a-real-ip%2F..%2Fetc" {
		t.Errorf("escaped path = %q, want escaped path segment", gotEscapedPath)
	}
}

func TestUpdateIPAddressOnlyIncludesNonNilFields(t *testing.T) {
	var gotBody map[string]any
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody = decodeBody(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ip":"1.2.3.4"}`))
	})

	isAutoWarmup := true
	if _, err := client.UpdateIPAddress(context.Background(), "1.2.3.4", &isAutoWarmup, nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(gotBody) != 1 || gotBody["is_auto_warmup"] != true {
		t.Errorf("body = %v, want only is_auto_warmup=true", gotBody)
	}
}

func TestAddAndRemoveSubusersFromIPAddress(t *testing.T) {
	var gotPaths []string
	var gotBodies []map[string]any
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.URL.Path)
		gotBodies = append(gotBodies, decodeBody(t, r))
		w.WriteHeader(http.StatusOK)
	})

	if err := client.AddSubusersToIPAddress(context.Background(), "1.2.3.4", []string{"1", "2"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := client.RemoveSubusersFromIPAddress(context.Background(), "1.2.3.4", []string{"1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPaths[0] != "/v3/send_ips/ips/1.2.3.4/subusers:batchAdd" {
		t.Errorf("path[0] = %q, want batchAdd path", gotPaths[0])
	}
	if gotPaths[1] != "/v3/send_ips/ips/1.2.3.4/subusers:batchDelete" {
		t.Errorf("path[1] = %q, want batchDelete path", gotPaths[1])
	}
}

func TestListIPAddressSubusers(t *testing.T) {
	var gotPath string
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"subusers":["1","2"]}`))
	})

	subs, err := client.ListIPAddressSubusers(context.Background(), "1.2.3.4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/v3/send_ips/ips/1.2.3.4/subusers" {
		t.Errorf("path = %q, want /v3/send_ips/ips/1.2.3.4/subusers", gotPath)
	}
	if len(subs) != 2 || subs[0] != "1" || subs[1] != "2" {
		t.Errorf("subs = %v, want [1 2]", subs)
	}
}
