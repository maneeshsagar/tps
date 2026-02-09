package currency

import "testing"

func TestRupeesToPaise(t *testing.T) {
	cases := []struct {
		in   string
		want int64
	}{
		{"100", 10000},
		{"100.50", 10050},
		{"100.5", 10050},
		{"0.50", 50},
		{"0.01", 1},
		{"0", 0},
	}

	for _, tc := range cases {
		got, err := RupeesToPaise(tc.in)
		if err != nil {
			t.Errorf("RupeesToPaise(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Errorf("RupeesToPaise(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestRupeesToPaise_Invalid(t *testing.T) {
	bad := []string{"", "abc", "12.3.4"}
	for _, s := range bad {
		if _, err := RupeesToPaise(s); err == nil {
			t.Errorf("RupeesToPaise(%q) should fail", s)
		}
	}
}

func TestPaiseToRupees(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{10000, "100.00"},
		{10050, "100.50"},
		{50, "0.50"},
		{1, "0.01"},
		{0, "0.00"},
	}

	for _, tc := range cases {
		if got := PaiseToRupees(tc.in); got != tc.want {
			t.Errorf("PaiseToRupees(%d) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
