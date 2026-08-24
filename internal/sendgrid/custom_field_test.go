package sendgrid

import (
	"context"
	"net/http"
	"testing"
)

func TestGetCustomFieldExactMatch(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"custom_fields": [{"id": "w1", "name": "num_orders", "field_type": "Number"}],
			"reserved_fields": [{"name": "first_name", "field_type": "Text"}]
		}`))
	})

	f, err := client.GetCustomField(context.Background(), "w1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f == nil || f.Name != "num_orders" || f.FieldType != "Number" {
		t.Errorf("got %+v, want id=w1 name=num_orders field_type=Number", f)
	}
}

func TestGetCustomFieldNotFound(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"custom_fields": [], "reserved_fields": []}`))
	})

	f, err := client.GetCustomField(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f != nil {
		t.Errorf("expected nil for no match, got %+v", f)
	}
}
