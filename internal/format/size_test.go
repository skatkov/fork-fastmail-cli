package format

import "testing"

func TestFormatSize(t *testing.T) {
	cases := []struct {
		name string
		in   int64
		want string
	}{
		{"bytes", 0, "0 B"},
		{"bytes-max", 1023, "1023 B"},
		{"kb", 1024, "1.0 KB"},
		{"mb", 1024 * 1024, "1.0 MB"},
		{"gb", 1024 * 1024 * 1024, "1.0 GB"},
		{"kb-round", 1536, "1.5 KB"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatSize(tt.in); got != tt.want {
				t.Fatalf("FormatSize(%d) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	cases := []struct {
		name string
		in   int64
		want string
	}{
		{"zero", 0, "0 B"},
		{"bytes", 1, "1 B"},
		{"kb", 1024, "1.0 KB"},
		{"mb", 1024 * 1024, "1.0 MB"},
		{"gb", 1024 * 1024 * 1024, "1.0 GB"},
		{"tb", 1024 * 1024 * 1024 * 1024, "1.0 TB"},
		{"tb-cap", 5 * 1024 * 1024 * 1024 * 1024, "5.0 TB"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatBytes(tt.in); got != tt.want {
				t.Fatalf("FormatBytes(%d) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
