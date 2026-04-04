package cmd

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// truncateCell
// ---------------------------------------------------------------------------

func TestTruncateCell(t *testing.T) {
	// Short value – no truncation.
	if got := truncateCell("hello", 20); got != "hello" {
		t.Errorf("truncateCell short = %q, want %q", got, "hello")
	}

	// Exact boundary – no truncation.
	if got := truncateCell("hello", 5); got != "hello" {
		t.Errorf("truncateCell exact = %q, want %q", got, "hello")
	}

	// Exceeds max – should truncate and add ellipsis.
	got := truncateCell("hello world", 7)
	if !strings.HasSuffix(got, "…") {
		t.Errorf("truncateCell long = %q, expected trailing '…'", got)
	}

	// maxWidth = 1 → just ellipsis.
	if got := truncateCell("hello", 1); got != "…" {
		t.Errorf("truncateCell maxWidth=1 = %q, want '…'", got)
	}
}

// ---------------------------------------------------------------------------
// padCell
// ---------------------------------------------------------------------------

func TestPadCell(t *testing.T) {
	// Exact width – no padding.
	if got := padCell("hi", 2); got != "hi" {
		t.Errorf("padCell exact = %q, want 'hi'", got)
	}

	// Pad to wider width.
	got := padCell("hi", 5)
	if len(got) != 5 {
		t.Errorf("padCell len = %d, want 5", len(got))
	}
	if !strings.HasPrefix(got, "hi") {
		t.Errorf("padCell result = %q, expected 'hi' prefix", got)
	}

	// Already wider than target – no truncation.
	if got := padCell("hello world", 3); got != "hello world" {
		t.Errorf("padCell wider = %q, want 'hello world'", got)
	}
}

// ---------------------------------------------------------------------------
// padStyledCell
// ---------------------------------------------------------------------------

func TestPadStyledCell(t *testing.T) {
	// Plain (no ANSI) strings behave like padCell.
	got := padStyledCell("hi", 5)
	if !strings.HasPrefix(got, "hi") {
		t.Errorf("padStyledCell = %q, expected 'hi' prefix", got)
	}
}

// ---------------------------------------------------------------------------
// renderAlignedTable
// ---------------------------------------------------------------------------

func TestRenderAlignedTable(t *testing.T) {
	headers := []string{"NAME", "STATUS"}
	rows := [][]string{
		{"default", "ready"},
		{"work", "stale"},
	}
	result := renderAlignedTable(headers, rows, nil, nil)

	for _, want := range []string{"NAME", "STATUS", "default", "work", "ready", "stale"} {
		if !strings.Contains(result, want) {
			t.Errorf("expected %q in table output: %q", want, result)
		}
	}
}

func TestRenderAlignedTableTruncatesWideCell(t *testing.T) {
	headers := []string{"NAME"}
	longName := strings.Repeat("a", 30)
	rows := [][]string{{longName}}
	result := renderAlignedTable(headers, rows, map[int]int{0: 10}, nil)
	if strings.Contains(result, longName) {
		t.Errorf("expected wide cell to be truncated, but found it intact in: %q", result)
	}
	if !strings.Contains(result, "…") {
		t.Errorf("expected truncation ellipsis in: %q", result)
	}
}

func TestRenderAlignedTableEmpty(t *testing.T) {
	result := renderAlignedTable([]string{"COL"}, nil, nil, nil)
	if !strings.Contains(result, "COL") {
		t.Errorf("expected header in empty table: %q", result)
	}
}
