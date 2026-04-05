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
			p:    NewTestProfile().WithHomebrew([]string{"git"}).Build(),
			want: 1,
		},
		{
			name: "multiple active sections",
			p: func() *profile.Profile {
				p := NewTestProfile().WithHomebrew([]string{"git"}).WithGit("Kasper", "kasper@example.com").Build()
				p.Languages.NodeVersion = "v20.0.0"
				return p
			}(),
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
