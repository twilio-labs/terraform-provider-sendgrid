package sendgrid

import "testing"

func TestTemplateVersionIsActive(t *testing.T) {
	one, zero := 1, 0

	tests := []struct {
		name   string
		active *int
		want   bool
	}{
		{name: "nil is not active", active: nil, want: false},
		{name: "0 is not active", active: &zero, want: false},
		{name: "1 is active", active: &one, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := TemplateVersion{Active: tt.active}
			if got := v.IsActive(); got != tt.want {
				t.Errorf("IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}
