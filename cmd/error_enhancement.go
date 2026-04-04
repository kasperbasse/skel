package cmd

import (
	"fmt"
	"strings"
)

// commandInstallHints keeps command-specific install guidance in one place.
var commandInstallHints = map[string]string{
	"brew":   "Install Homebrew first: https://brew.sh",
	"gh":     "Install GitHub CLI: https://cli.github.com",
	"mas":    "Install mas for Mac App Store: brew install mas",
	"code":   "Install VS Code or ensure it's in your PATH",
	"cursor": "Install Cursor or ensure it's in your PATH",
}

type errorEnhancementRule struct {
	match   func(errMsg string) bool
	enhance func(errMsg string) error
}

var errorEnhancementRules = []errorEnhancementRule{
	{
		match: func(errMsg string) bool {
			return strings.Contains(errMsg, "profile") && strings.Contains(errMsg, "not found")
		},
		enhance: func(errMsg string) error {
			profileName := extractProfileName(errMsg)
			if profileName != "" {
				suggestion := suggestSimilarProfile(profileName)
				if suggestion != "" {
					return fmt.Errorf("%s\n\nDid you mean '%s'? Use 'skel list' to see all profiles", errMsg, suggestion)
				}
			}
			return fmt.Errorf("%s\n\nUse 'skel list' to see available profiles", errMsg)
		},
	},
	{
		match: func(errMsg string) bool {
			return strings.Contains(errMsg, "unknown section") || strings.Contains(errMsg, "invalid section")
		},
		enhance: func(errMsg string) error {
			return fmt.Errorf("%s\n\nValid sections: %s", errMsg, cyan(strings.Join(allRestoreKeys(), ", ")))
		},
	},
	{
		match: func(errMsg string) bool {
			return strings.Contains(errMsg, "not found")
		},
		enhance: func(errMsg string) error {
			cmd := extractCommandName(errMsg)
			if cmd == "" {
				return nil
			}
			hint, ok := commandInstallHints[cmd]
			if !ok {
				return nil
			}
			return fmt.Errorf("%s\n\n%s", errMsg, hint)
		},
	},
	{
		match: func(errMsg string) bool {
			return strings.Contains(errMsg, "json") && containsAny(errMsg, "invalid", "parsing", "unmarshal")
		},
		enhance: func(errMsg string) error {
			return fmt.Errorf("%s\n\nCheck that the file contains valid JSON. For profiles, use 'skel export' to create valid files", errMsg)
		},
	},
	{
		match: func(errMsg string) bool {
			return strings.Contains(errMsg, "rate limit")
		},
		enhance: func(errMsg string) error {
			return fmt.Errorf("%s\n\nSet GITHUB_TOKEN environment variable or run 'gh auth login' to increase limits", errMsg)
		},
	},
	{
		match: func(errMsg string) bool {
			return strings.Contains(errMsg, "authentication failed") || strings.Contains(errMsg, "401")
		},
		enhance: func(errMsg string) error {
			return fmt.Errorf("%s\n\nSet GITHUB_TOKEN or run 'gh auth login' to authenticate", errMsg)
		},
	},
	{
		match: func(errMsg string) bool {
			return containsAny(errMsg, "too large", "size", "max")
		},
		enhance: func(errMsg string) error {
			if strings.Contains(errMsg, "profile") {
				return fmt.Errorf("%s\n\nTry reducing config file sizes or split into multiple profiles", errMsg)
			}
			return fmt.Errorf("%s\n\nTry a smaller file or contact support if this is unexpected", errMsg)
		},
	},
	{
		match: func(errMsg string) bool {
			return containsAny(errMsg, "permission denied", "access denied")
		},
		enhance: func(errMsg string) error {
			return fmt.Errorf("%s\n\nTry running with %s if you have admin privileges", errMsg, cyan("sudo"))
		},
	},
	{
		match: func(errMsg string) bool {
			return containsAny(errMsg, "connection refused", "network", "timeout")
		},
		enhance: func(errMsg string) error {
			return fmt.Errorf("%s\n\nCheck your internet connection and try again", errMsg)
		},
	},
	{
		match: func(errMsg string) bool {
			return strings.Contains(errMsg, "github") || strings.Contains(errMsg, "gist")
		},
		enhance: func(errMsg string) error {
			if strings.Contains(errMsg, "token") {
				return fmt.Errorf("%s\n\nSet GITHUB_TOKEN or run 'gh auth login' to authenticate", errMsg)
			}
			return fmt.Errorf("%s\n\nCheck the gist URL and your GitHub authentication", errMsg)
		},
	},
}

func applyErrorEnhancementRules(errMsg string) error {
	for _, rule := range errorEnhancementRules {
		if !rule.match(errMsg) {
			continue
		}
		enhanced := rule.enhance(errMsg)
		if enhanced != nil {
			return enhanced
		}
	}
	return nil
}

func containsAny(value string, candidates ...string) bool {
	for _, candidate := range candidates {
		if strings.Contains(value, candidate) {
			return true
		}
	}
	return false
}
