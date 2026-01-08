package format

import "testing"

func TestTruncate(t *testing.T) {
	cases := []struct {
		name   string
		in     string
		maxLen int
		want   string
	}{
		{"short", "hello", 10, "hello"},
		{"exact", "hello", 5, "hello"},
		{"long", "hello", 4, "h..."},
		{"longer", "abcdefghij", 6, "abc..."},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := Truncate(tt.in, tt.maxLen); got != tt.want {
				t.Fatalf("Truncate(%q, %d) = %q, want %q", tt.in, tt.maxLen, got, tt.want)
			}
		})
	}
}
