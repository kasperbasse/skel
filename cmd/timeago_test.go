package cmd

import (
	"testing"
	"time"
)

func TestTimeAgo(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name string
		t    time.Time
		want string
	}{
		{"just now", now.Add(-30 * time.Second), "just now"},
		{"1 minute ago", now.Add(-90 * time.Second), "1 minute ago"},
		{"5 minutes ago", now.Add(-5 * time.Minute), "5 minutes ago"},
		{"59 minutes ago", now.Add(-59 * time.Minute), "59 minutes ago"},
		{"1 hour ago", now.Add(-90 * time.Minute), "1 hour ago"},
		{"3 hours ago", now.Add(-3 * time.Hour), "3 hours ago"},
		{"23 hours ago", now.Add(-23 * time.Hour), "23 hours ago"},
		{"yesterday", now.Add(-36 * time.Hour), "yesterday"},
		{"3 days ago", now.Add(-3 * 24 * time.Hour), "3 days ago"},
		{"6 days ago", now.Add(-6 * 24 * time.Hour), "6 days ago"},
		{"1 week ago", now.Add(-10 * 24 * time.Hour), "1 week ago"},
		{"2 weeks ago", now.Add(-16 * 24 * time.Hour), "2 weeks ago"},
		{"3 weeks ago", now.Add(-23 * 24 * time.Hour), "3 weeks ago"},
		{"1 month ago", now.Add(-45 * 24 * time.Hour), "1 month ago"},
		{"3 months ago", now.Add(-100 * 24 * time.Hour), "3 months ago"},
		{"11 months ago", now.Add(-350 * 24 * time.Hour), "11 months ago"},
		{"1 year ago", now.Add(-400 * 24 * time.Hour), "1 year ago"},
		{"2 years ago", now.Add(-800 * 24 * time.Hour), "2 years ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := timeAgo(tt.t)
			if got != tt.want {
				t.Errorf("timeAgo(%v) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestTimeAgoSingularPlural(t *testing.T) {
	now := time.Now()

	// Verify singular forms don't say "1 minutes ago" etc.
	singulars := []struct {
		d    time.Duration
		want string
	}{
		{90 * time.Second, "1 minute ago"},
		{90 * time.Minute, "1 hour ago"},
		{10 * 24 * time.Hour, "1 week ago"},
		{45 * 24 * time.Hour, "1 month ago"},
		{400 * 24 * time.Hour, "1 year ago"},
	}
	for _, s := range singulars {
		got := timeAgo(now.Add(-s.d))
		if got != s.want {
			t.Errorf("singular: timeAgo(-%v) = %q, want %q", s.d, got, s.want)
		}
	}
}
