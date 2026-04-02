package brewfile

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/kasperbasse/skel/internal/profile"
)

// MaxBrewfileSize is the maximum size of a Brewfile we'll parse (1 MB).
const MaxBrewfileSize = 1024 * 1024

// maxEntries prevents resource exhaustion from absurdly large Brewfiles.
const maxEntries = 10000

// validName matches safe Homebrew package names: alphanumeric, dashes, underscores, slashes, dots, @.
var validName = regexp.MustCompile(`^[a-zA-Z0-9_@/.+-]+$`)

// Generate produces a standard Brewfile from a HomebrewProfile.
func Generate(h profile.HomebrewProfile) string {
	var b strings.Builder

	for _, t := range h.Taps {
		_, _ = fmt.Fprintf(&b, "tap %q\n", t)
	}
	if len(h.Taps) > 0 && (len(h.Formulas) > 0 || len(h.Casks) > 0 || len(h.MasApps) > 0) {
		b.WriteString("\n")
	}

	for _, f := range h.Formulas {
		_, _ = fmt.Fprintf(&b, "brew %q\n", f)
	}
	if len(h.Formulas) > 0 && (len(h.Casks) > 0 || len(h.MasApps) > 0) {
		b.WriteString("\n")
	}

	for _, c := range h.Casks {
		_, _ = fmt.Fprintf(&b, "cask %q\n", c)
	}
	if len(h.Casks) > 0 && len(h.MasApps) > 0 {
		b.WriteString("\n")
	}

	for _, app := range h.MasApps {
		_, _ = fmt.Fprintf(&b, "mas %q, id: %s\n", app.Name, app.ID)
	}

	return b.String()
}

// Parse reads a Brewfile and returns a HomebrewProfile.
// It validates package names to prevent shell injection.
func Parse(content string) (profile.HomebrewProfile, error) {
	if len(content) > MaxBrewfileSize {
		return profile.HomebrewProfile{}, fmt.Errorf("brewfile too large (%d bytes, max %d)", len(content), MaxBrewfileSize)
	}

	h := profile.HomebrewProfile{}
	total := 0

	for i, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		total++
		if total > maxEntries {
			return profile.HomebrewProfile{}, fmt.Errorf("brewfile exceeds maximum of %d entries", maxEntries)
		}

		lineNum := i + 1

		switch {
		case strings.HasPrefix(line, "tap "):
			name, err := extractQuotedOrBare(line[4:])
			if err != nil {
				return profile.HomebrewProfile{}, fmt.Errorf("line %d: invalid tap: %w", lineNum, err)
			}
			if err := validateName(name); err != nil {
				return profile.HomebrewProfile{}, fmt.Errorf("line %d: %w", lineNum, err)
			}
			h.Taps = append(h.Taps, name)

		case strings.HasPrefix(line, "brew "):
			name, err := extractQuotedOrBare(line[5:])
			if err != nil {
				return profile.HomebrewProfile{}, fmt.Errorf("line %d: invalid brew: %w", lineNum, err)
			}
			if err := validateName(name); err != nil {
				return profile.HomebrewProfile{}, fmt.Errorf("line %d: %w", lineNum, err)
			}
			h.Formulas = append(h.Formulas, name)

		case strings.HasPrefix(line, "cask "):
			name, err := extractQuotedOrBare(line[5:])
			if err != nil {
				return profile.HomebrewProfile{}, fmt.Errorf("line %d: invalid cask: %w", lineNum, err)
			}
			if err := validateName(name); err != nil {
				return profile.HomebrewProfile{}, fmt.Errorf("line %d: %w", lineNum, err)
			}
			h.Casks = append(h.Casks, name)

		case strings.HasPrefix(line, "mas "):
			app, err := parseMasLine(line)
			if err != nil {
				return profile.HomebrewProfile{}, fmt.Errorf("line %d: %w", lineNum, err)
			}
			h.MasApps = append(h.MasApps, app)

		default:
			return profile.HomebrewProfile{}, fmt.Errorf("line %d: unknown directive %q", lineNum, strings.Fields(line)[0])
		}
	}

	return h, nil
}

// extractQuotedOrBare extracts a package name from either "quoted" or bare format.
// It strips trailing options like ", restart_service: true".
func extractQuotedOrBare(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("empty value")
	}

	if s[0] == '"' {
		end := strings.Index(s[1:], "\"")
		if end < 0 {
			return "", fmt.Errorf("unterminated quote")
		}
		return s[1 : end+1], nil
	}

	// Bare format: take everything up to comma or whitespace
	name := s
	if idx := strings.IndexAny(name, ", \t"); idx >= 0 {
		name = name[:idx]
	}
	return name, nil
}

// parseMasLine parses: mas "App Name", id: 123456
func parseMasLine(line string) (profile.MasApp, error) {
	rest := strings.TrimSpace(line[4:])

	name, err := extractQuotedOrBare(rest)
	if err != nil {
		return profile.MasApp{}, fmt.Errorf("invalid mas app name: %w", err)
	}

	// Find id: NNN
	idIdx := strings.Index(rest, "id:")
	if idIdx < 0 {
		return profile.MasApp{}, fmt.Errorf("mas line missing 'id:' field")
	}

	idStr := strings.TrimSpace(rest[idIdx+3:])
	// Take only digits
	idEnd := 0
	for idEnd < len(idStr) && idStr[idEnd] >= '0' && idStr[idEnd] <= '9' {
		idEnd++
	}
	if idEnd == 0 {
		return profile.MasApp{}, fmt.Errorf("mas line has non-numeric id")
	}
	idVal := idStr[:idEnd]

	// Validate it's actually a number
	if _, err := strconv.Atoi(idVal); err != nil {
		return profile.MasApp{}, fmt.Errorf("mas id is not a valid number: %s", idVal)
	}

	return profile.MasApp{
		Name: name,
		ID:   idVal,
	}, nil
}

func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("empty package name")
	}
	if !validName.MatchString(name) {
		return fmt.Errorf("unsafe package name %q (contains disallowed characters)", name)
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("package name %q contains path traversal", name)
	}
	return nil
}
