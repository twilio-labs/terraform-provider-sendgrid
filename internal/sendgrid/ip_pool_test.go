package sendgrid

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

// TestClientGetIPPoolDecodesIPObjects locks in a real API response shape:
// GET /v3/ips/pools/{pool_name} returns "ips" as an array of {ip, start_date,
// warmup} objects, not a flat array of strings. A prior version of this
// client assumed a flat []string and failed to decode any pool with at
// least one IP assigned, confirmed against a live account.
func TestClientGetIPPoolDecodesIPObjects(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{
			"pool_name": "test-pool",
			"ips": [
				{"ip": "203.0.113.10", "start_date": null, "warmup": false}
			]
		}`)
	})

	pool, err := client.GetIPPool(context.Background(), "test-pool")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pool.Name != "test-pool" {
		t.Errorf("Name = %q, want %q", pool.Name, "test-pool")
	}
	if want := []string{"203.0.113.10"}; len(pool.IPs) != len(want) || pool.IPs[0] != want[0] {
		t.Errorf("IPs = %v, want %v", pool.IPs, want)
	}
}
