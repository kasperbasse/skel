package errorx

import "testing"

func TestExtractProfileName(t *testing.T) {
	tests := []struct {
		msg  string
		want string
	}{
		{"profile 'work' not found", "work"},
		{"profile \"work\" not found", "work"},
		{"profile default: no such file", "default"},
		{"unknown error", ""},
	}

	for _, tt := range tests {
		if got := ExtractProfileName(tt.msg); got != tt.want {
			t.Fatalf("ExtractProfileName(%q) = %q, want %q", tt.msg, got, tt.want)
		}
	}
}

func TestExtractCommandName(t *testing.T) {
	tests := []struct {
		msg  string
		want string
	}{
		{"gh: command not found", "gh"},
		{"brew not found", "brew"},
		{"executable file not found in $PATH: mas", "mas"},
		{"some other error", ""},
	}

	for _, tt := range tests {
		if got := ExtractCommandName(tt.msg); got != tt.want {
			t.Fatalf("ExtractCommandName(%q) = %q, want %q", tt.msg, got, tt.want)
		}
	}
}

func TestSuggestClosestName(t *testing.T) {
	candidates := []string{"default", "work", "personal"}
	if got := SuggestClosestName("wrk", candidates, 2); got != "work" {
		t.Fatalf("SuggestClosestName() = %q, want %q", got, "work")
	}
	if got := SuggestClosestName("zzzz", candidates, 2); got != "" {
		t.Fatalf("SuggestClosestName() = %q, want empty", got)
	}
}
