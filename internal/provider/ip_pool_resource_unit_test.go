package provider

import (
	"reflect"
	"sort"
	"testing"
)

func TestDiffIPPool(t *testing.T) {
	tests := []struct {
		name       string
		current    []string
		want       []string
		wantAdd    []string
		wantRemove []string
	}{
		{
			name:    "no changes",
			current: []string{"1.1.1.1", "2.2.2.2"},
			want:    []string{"1.1.1.1", "2.2.2.2"},
		},
		{
			name:    "pure additions",
			current: []string{"1.1.1.1"},
			want:    []string{"1.1.1.1", "2.2.2.2"},
			wantAdd: []string{"2.2.2.2"},
		},
		{
			name:       "pure removals",
			current:    []string{"1.1.1.1", "2.2.2.2"},
			want:       []string{"1.1.1.1"},
			wantRemove: []string{"2.2.2.2"},
		},
		{
			name:       "mixed add and remove",
			current:    []string{"1.1.1.1", "2.2.2.2"},
			want:       []string{"2.2.2.2", "3.3.3.3"},
			wantAdd:    []string{"3.3.3.3"},
			wantRemove: []string{"1.1.1.1"},
		},
		{
			name:    "duplicate entries in input are not duplicated in output",
			current: []string{"1.1.1.1", "1.1.1.1"},
			want:    []string{"1.1.1.1"},
		},
		{
			name:    "empty current",
			current: nil,
			want:    []string{"1.1.1.1"},
			wantAdd: []string{"1.1.1.1"},
		},
		{
			name:       "empty want",
			current:    []string{"1.1.1.1"},
			want:       nil,
			wantRemove: []string{"1.1.1.1"},
		},
		{
			name:    "reorder only produces no changes",
			current: []string{"1.1.1.1", "2.2.2.2", "3.3.3.3"},
			want:    []string{"3.3.3.3", "1.1.1.1", "2.2.2.2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAdd, gotRemove := diffIPPool(tt.current, tt.want)
			assertSameElements(t, "toAdd", gotAdd, tt.wantAdd)
			assertSameElements(t, "toRemove", gotRemove, tt.wantRemove)
		})
	}
}

// assertSameElements compares two string slices as sets, treating nil and
// an empty slice as equal (diffIPPool returns nil for "no changes", but a
// test case's zero-value want field is also nil, so a plain length+sort
// comparison avoids reflect.DeepEqual's nil-vs-empty-slice distinction).
func assertSameElements(t *testing.T, label string, got, want []string) {
	t.Helper()
	gotSorted := append([]string{}, got...)
	wantSorted := append([]string{}, want...)
	sort.Strings(gotSorted)
	sort.Strings(wantSorted)
	if !reflect.DeepEqual(gotSorted, wantSorted) {
		t.Errorf("%s = %v, want %v", label, got, want)
	}
}
