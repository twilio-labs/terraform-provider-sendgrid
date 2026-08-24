package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestIDFromInt64(t *testing.T) {
	tests := []struct {
		id   int64
		want string
	}{
		{id: 0, want: "0"},
		{id: 42, want: "42"},
		{id: -1, want: "-1"},
	}
	for _, tt := range tests {
		if got := idFromInt64(tt.id); got != tt.want {
			t.Errorf("idFromInt64(%d) = %q, want %q", tt.id, got, tt.want)
		}
	}
}

func TestStringSetToSlice(t *testing.T) {
	ctx := context.Background()

	if got, diags := stringSetToSlice(ctx, types.SetNull(types.StringType)); len(got) != 0 || diags.HasError() {
		t.Errorf("null set: got %v, diags %v", got, diags)
	}
	if got, diags := stringSetToSlice(ctx, types.SetUnknown(types.StringType)); len(got) != 0 || diags.HasError() {
		t.Errorf("unknown set: got %v, diags %v", got, diags)
	}

	set, diags := types.SetValueFrom(ctx, types.StringType, []string{"a", "b"})
	if diags.HasError() {
		t.Fatalf("building test set: %v", diags)
	}
	got, diags := stringSetToSlice(ctx, set)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("got %v, want [a b]", got)
	}
}

func TestSetStringSetRoundTrip(t *testing.T) {
	ctx := context.Background()

	var dst types.Set
	if diags := setStringSet(ctx, &dst, nil); diags.HasError() {
		t.Fatalf("unexpected error setting nil: %v", diags)
	}
	got, diags := stringSetToSlice(ctx, dst)
	if diags.HasError() || len(got) != 0 {
		t.Errorf("round-tripping nil: got %v, diags %v", got, diags)
	}

	if diags := setStringSet(ctx, &dst, []string{"x", "y"}); diags.HasError() {
		t.Fatalf("unexpected error setting values: %v", diags)
	}
	got, diags = stringSetToSlice(ctx, dst)
	if diags.HasError() || len(got) != 2 {
		t.Errorf("round-tripping [x y]: got %v, diags %v", got, diags)
	}
}

func TestStringListToSlice(t *testing.T) {
	ctx := context.Background()

	if got, diags := stringListToSlice(ctx, types.ListNull(types.StringType)); len(got) != 0 || diags.HasError() {
		t.Errorf("null list: got %v, diags %v", got, diags)
	}
	if got, diags := stringListToSlice(ctx, types.ListUnknown(types.StringType)); len(got) != 0 || diags.HasError() {
		t.Errorf("unknown list: got %v, diags %v", got, diags)
	}

	list, diags := types.ListValueFrom(ctx, types.StringType, []string{"a", "b"})
	if diags.HasError() {
		t.Fatalf("building test list: %v", diags)
	}
	got, diags := stringListToSlice(ctx, list)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("got %v, want [a b]", got)
	}
}

func TestSetStringListRoundTrip(t *testing.T) {
	ctx := context.Background()

	var dst types.List
	if diags := setStringList(ctx, &dst, nil); diags.HasError() {
		t.Fatalf("unexpected error setting nil: %v", diags)
	}
	got, diags := stringListToSlice(ctx, dst)
	if diags.HasError() || len(got) != 0 {
		t.Errorf("round-tripping nil: got %v, diags %v", got, diags)
	}

	if diags := setStringList(ctx, &dst, []string{"x", "y"}); diags.HasError() {
		t.Fatalf("unexpected error setting values: %v", diags)
	}
	got, diags = stringListToSlice(ctx, dst)
	if diags.HasError() || len(got) != 2 {
		t.Errorf("round-tripping [x y]: got %v, diags %v", got, diags)
	}
}
