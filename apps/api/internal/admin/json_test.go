package admin

import "testing"

func TestParsePositiveInt(t *testing.T) {
	cases := []struct {
		raw      string
		fallback int
		want     int
	}{
		{"", 10, 10},
		{"5", 10, 5},
		{"0", 10, 10},
		{"-3", 10, 10},
		{"not-a-number", 10, 10},
		{"42", 1, 42},
	}

	for _, c := range cases {
		got := parsePositiveInt(c.raw, c.fallback)
		if got != c.want {
			t.Fatalf("parsePositiveInt(%q, %d) = %d, want %d", c.raw, c.fallback, got, c.want)
		}
	}
}
