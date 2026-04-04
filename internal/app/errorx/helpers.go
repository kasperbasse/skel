package errorx

import (
	"regexp"
	"strings"

	"github.com/kasperbasse/skel/internal/profile"
)

var profileNamePatterns = []*regexp.Regexp{
	regexp.MustCompile(`profile '([^']+)'`),
	regexp.MustCompile(`profile ([^:]+):`),
	regexp.MustCompile(`([^\s]+): no such file`),
}

var commandNamePatterns = []*regexp.Regexp{
	regexp.MustCompile(`executable file not found in \$PATH: (\w+)`),
	regexp.MustCompile(`(\w+): command not found`),
	regexp.MustCompile(`(\w+) not found`),
}

// ExtractProfileName returns a best-effort profile name from an error string.
func ExtractProfileName(errMsg string) string {
	for _, re := range profileNamePatterns {
		matches := re.FindStringSubmatch(errMsg)
		if len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}

// ExtractCommandName returns a best-effort command name from an error string.
func ExtractCommandName(errMsg string) string {
	for _, re := range commandNamePatterns {
		matches := re.FindStringSubmatch(errMsg)
		if len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}

// SuggestSimilarProfileName suggests the closest saved profile name within edit distance 2.
func SuggestSimilarProfileName(wrongName string) string {
	profiles, err := profile.ListAll()
	if err != nil {
		return ""
	}

	names := make([]string, 0, len(profiles))
	for _, p := range profiles {
		names = append(names, p.Name)
	}

	return SuggestClosestName(wrongName, names, 2)
}

// SuggestClosestName returns the closest candidate within maxDistance.
func SuggestClosestName(wrongName string, candidates []string, maxDistance int) string {
	minDistance := maxDistance + 1
	bestMatch := ""
	for _, candidate := range candidates {
		distance := levenshteinDistance(strings.ToLower(wrongName), strings.ToLower(candidate))
		if distance < minDistance {
			minDistance = distance
			bestMatch = candidate
		}
	}
	return bestMatch
}

func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			matrix[i][j] = min3(
				matrix[i-1][j]+1,
				matrix[i][j-1]+1,
				matrix[i-1][j-1]+cost,
			)
		}
	}

	return matrix[len(a)][len(b)]
}

func min3(a, b, c int) int {
	if a < b && a < c {
		return a
	}
	if b < c {
		return b
	}
	return c
}
