package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kasperbasse/skel/internal/profile"
)

func wrapIfNotEmpty(value string) []string {
	if value == "" {
		return nil
	}
	return []string{value}
}

func profileItemCount(p *profile.Profile) int {
	n := 0
	for _, s := range profileSections {
		n += len(s.Items(p))
	}
	return n
}

func profileSummaryParts(p *profile.Profile) []string {
	var parts []string
	for _, s := range profileSections {
		items := s.Items(p)
		if len(items) > 0 {
			parts = append(parts, num(len(items))+" "+s.Label)
		}
	}
	return parts
}

func printDiffSection(icon, title string, added, removed []string) {
	count := len(added) + len(removed)
	fmt.Printf("  %s %s %s\n", icon, bold(title), dim(fmt.Sprintf("(%d)", count)))
	for _, f := range added {
		fmt.Printf("     %s %s\n", green("+"), green(f))
	}
	for _, f := range removed {
		fmt.Printf("     %s %s\n", red("-"), red(f))
	}
	fmt.Println()
}

func diffSlices(a, b []string) (added, removed []string) {
	setA := toSet(a)
	setB := toSet(b)
	for item := range setB {
		if !setA[item] {
			added = append(added, item)
		}
	}
	for item := range setA {
		if !setB[item] {
			removed = append(removed, item)
		}
	}
	return
}

func toSet(s []string) map[string]bool {
	m := make(map[string]bool, len(s))
	for _, v := range s {
		m[v] = true
	}
	return m
}

func summarizeVersions(p *profile.Profile) string {
	var parts []string
	dot := subtleStyle.Render(" · ")

	for _, v := range versionFields {
		ver := v.Value(p)
		if ver == "" {
			continue
		}

		cleanVer := ver
		if strings.Contains(ver, "go version") {
			cleanVer = strings.Split(strings.Replace(ver, "go version go", "", 1), " ")[0]
		} else if strings.Contains(ver, "PHP") || strings.Contains(ver, "ruby") || strings.Contains(ver, "Python") || strings.Contains(ver, "Rust") {
			fields := strings.Fields(ver)
			if len(fields) > 1 {
				cleanVer = fields[1]
			}
		} else if v.Label == "Java" {
			fields := strings.Fields(ver)
			for _, f := range fields {
				if strings.Contains(f, ".") {
					cleanVer = strings.Trim(f, "\"")
					break
				}
			}
		}

		parts = append(parts, fmt.Sprintf("%s %s", subtleStyle.Render(v.DisplayLabel), versionStyle.Render(cleanVer)))
	}

	pkgs := []struct {
		label string
		count int
	}{
		{"NPM", len(p.Languages.NpmGlobals)},
		{"Yarn", len(p.Languages.YarnGlobals)},
		{"PNPM", len(p.Languages.PnpmGlobals)},
		{"Pip", len(p.Languages.PipGlobals)},
		{"Composer", len(p.Languages.ComposerGlobals)},
		{"Ruby Gems", len(p.Languages.GemGlobals)},
		{"Cargo", len(p.Languages.CargoPackages)},
	}

	for _, pkg := range pkgs {
		if pkg.count > 0 {
			parts = append(parts, fmt.Sprintf("%s %s", countStyle.Render(strconv.Itoa(pkg.count)), subtleStyle.Render(pkg.label)))
		}
	}

	return strings.Join(parts, dot)
}
