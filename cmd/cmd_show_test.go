package cmd

import (
	"testing"

	"github.com/kasperbasse/skel/internal/profile"
)

func TestCountActiveSections(t *testing.T) {
	tests := []struct {
		name string
		p    *profile.Profile
		want int
	}{
		{
			name: "empty profile",
			p:    &profile.Profile{},
			want: 0,
		},
		{
			name: "single active section",
			p: &profile.Profile{
				Homebrew: profile.HomebrewProfile{Formulas: []string{"git"}},
			},
			want: 1,
		},
		{
			name: "multiple active sections",
			p: &profile.Profile{
				Homebrew: profile.HomebrewProfile{Formulas: []string{"git"}},
				Git: profile.GitProfile{
					UserName:  "Kasper",
					UserEmail: "kasper@example.com",
				},
				Languages: profile.LanguageProfile{NodeVersion: "v20.0.0"},
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := countActiveSections(tt.p); got != tt.want {
				t.Fatalf("countActiveSections() = %d, want %d", got, tt.want)
			}
		})
	}
}
