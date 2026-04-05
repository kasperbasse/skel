package cmd

import (
	"fmt"
	"strings"
)

func printProfileHeader(title, name string) {
	fmt.Printf("\n  %s %s: %s\n", cyan(headlineIcon(strings.ToLower(title))), bold(title), bold("'"+name+"'"))
	fmt.Printf("  %s\n", dividerStyle.Render(dividerLine))
}
