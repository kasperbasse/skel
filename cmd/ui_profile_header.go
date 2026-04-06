package cmd

import (
	"fmt"
	"strings"
)

func printProfileHeader(title, name string, subtitle ...string) {
	fmt.Printf("\n  %s %s: %s\n", cyan(headlineIcon(strings.ToLower(title))), bold(title), bold("'"+name+"'"))
	fmt.Printf("  %s\n", dividerStyle.Render(dividerLine))
	if len(subtitle) > 0 && subtitle[0] != "" {
		fmt.Printf("  %s\n", dim(subtitle[0]))
	}
}
