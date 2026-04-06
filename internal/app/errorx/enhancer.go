package errorx

import "strings"

// EnhanceOptions provides dependencies and runtime values needed by rule evaluation.
type EnhanceOptions struct {
	ValidSections  string
	ToolHint       func(command string) (string, bool)
	SuggestProfile func(wrongName string) string
}

// Rule matches an error message and optionally returns an enhanced replacement.
type Rule struct {
	Match   func(errMsg string) bool
	Enhance func(errMsg string, opts EnhanceOptions) (string, bool)
}

var defaultRules = buildDefaultRules()

func buildDefaultRules() []Rule {
	return []Rule{
		{
			Match: func(errMsg string) bool {
				return strings.Contains(errMsg, "profile") && strings.Contains(errMsg, "not found")
			},
			Enhance: func(errMsg string, opts EnhanceOptions) (string, bool) {
				profileName := ExtractProfileName(errMsg)
				if profileName != "" && opts.SuggestProfile != nil {
					suggestion := opts.SuggestProfile(profileName)
					if suggestion != "" {
						return errMsg + "\n\nDid you mean '" + suggestion + "'? Use 'skel list' to see all profiles", true
					}
				}
				return errMsg + "\n\n  Use 'skel list' to see available profiles", true
			},
		},
		{
			Match: func(errMsg string) bool {
				return strings.Contains(errMsg, "unknown section") || strings.Contains(errMsg, "invalid section")
			},
			Enhance: func(errMsg string, opts EnhanceOptions) (string, bool) {
				if opts.ValidSections == "" {
					return "", false
				}
				return errMsg + "\n\n  Valid sections: " + opts.ValidSections, true
			},
		},
		{
			Match: func(errMsg string) bool {
				return strings.Contains(errMsg, "not found")
			},
			Enhance: func(errMsg string, opts EnhanceOptions) (string, bool) {
				if opts.ToolHint == nil {
					return "", false
				}
				cmd := ExtractCommandName(errMsg)
				if cmd == "" {
					return "", false
				}
				hint, ok := opts.ToolHint(cmd)
				if !ok {
					return "", false
				}
				return errMsg + "\n\n  " + hint, true
			},
		},
		{
			Match: func(errMsg string) bool {
				return strings.Contains(errMsg, "json") && containsAny(errMsg, "invalid", "parsing", "unmarshal")
			},
			Enhance: func(errMsg string, _ EnhanceOptions) (string, bool) {
				return errMsg + "\n\n  Check that the file contains valid JSON. For profiles, use 'skel export' to create valid files", true
			},
		},
		{
			Match: func(errMsg string) bool { return strings.Contains(errMsg, "rate limit") },
			Enhance: func(errMsg string, _ EnhanceOptions) (string, bool) {
				return errMsg + "\n\n  Set GITHUB_TOKEN environment variable or run 'gh auth login' to increase limits", true
			},
		},
		{
			Match: func(errMsg string) bool {
				return strings.Contains(errMsg, "authentication failed") || strings.Contains(errMsg, "401")
			},
			Enhance: func(errMsg string, _ EnhanceOptions) (string, bool) {
				return errMsg + "\n\n  Set GITHUB_TOKEN or run 'gh auth login' to authenticate", true
			},
		},
		{
			Match: func(errMsg string) bool { return containsAny(errMsg, "too large", "size", "max") },
			Enhance: func(errMsg string, _ EnhanceOptions) (string, bool) {
				if strings.Contains(errMsg, "profile") {
					return errMsg + "\n\n  Try reducing config file sizes or split into multiple profiles", true
				}
				return errMsg + "\n\n  Try a smaller file or contact support if this is unexpected", true
			},
		},
		{
			Match: func(errMsg string) bool { return containsAny(errMsg, "permission denied", "access denied") },
			Enhance: func(errMsg string, _ EnhanceOptions) (string, bool) {
				return errMsg + "\n\n  Try running with sudo if you have admin privileges", true
			},
		},
		{
			Match: func(errMsg string) bool { return containsAny(errMsg, "connection refused", "network", "timeout") },
			Enhance: func(errMsg string, _ EnhanceOptions) (string, bool) {
				return errMsg + "\n\n  Check your internet connection and try again", true
			},
		},
		{
			Match: func(errMsg string) bool {
				return strings.Contains(errMsg, "github") || strings.Contains(errMsg, "gist")
			},
			Enhance: func(errMsg string, _ EnhanceOptions) (string, bool) {
				if strings.Contains(errMsg, "token") {
					return errMsg + "\n\n  Set GITHUB_TOKEN or run 'gh auth login' to authenticate", true
				}
				return errMsg + "\n\n  Check the gist URL and your GitHub authentication", true
			},
		},
	}
}

func EnhanceMessage(errMsg string, opts EnhanceOptions) (string, bool) {
	for _, rule := range defaultRules {
		if !rule.Match(errMsg) {
			continue
		}
		if msg, ok := rule.Enhance(errMsg, opts); ok {
			return msg, true
		}
	}
	return "", false
}

func containsAny(value string, candidates ...string) bool {
	for _, candidate := range candidates {
		if strings.Contains(value, candidate) {
			return true
		}
	}
	return false
}
