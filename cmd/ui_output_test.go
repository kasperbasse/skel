package cmd

import "testing"

func TestIsAffirmative(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: "y", in: "y", want: true},
		{name: "yes", in: "yes", want: true},
		{name: "uppercase", in: "Y", want: true},
		{name: "spaced", in: "  yes  ", want: true},
		{name: "no", in: "n", want: false},
		{name: "empty", in: "", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isAffirmative(tc.in); got != tc.want {
				t.Fatalf("isAffirmative(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}
