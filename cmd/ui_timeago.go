package cmd

import (
	"fmt"
	"time"
)

// timeAgo returns a human-readable relative duration string for a past time,
// e.g. "just now", "3 hours ago", "2 days ago", "1 week ago".
func timeAgo(t time.Time) string {
	d := time.Since(t)

	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	case d < 2*24*time.Hour:
		return "yesterday"
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%d days ago", int(d.Hours()/24))
	case d < 14*24*time.Hour:
		return "1 week ago"
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%d weeks ago", int(d.Hours()/(24*7)))
	case d < 60*24*time.Hour:
		return "1 month ago"
	case d < 365*24*time.Hour:
		return fmt.Sprintf("%d months ago", int(d.Hours()/(24*30)))
	case d < 2*365*24*time.Hour:
		return "1 year ago"
	default:
		return fmt.Sprintf("%d years ago", int(d.Hours()/(24*365)))
	}
}
