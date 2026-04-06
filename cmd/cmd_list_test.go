package cmd

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kasperbasse/skel/cmd/tui"
	"github.com/kasperbasse/skel/internal/profile"
)

type fakeTeaModel struct{}

func (fakeTeaModel) Init() tea.Cmd                              { return nil }
func (fakeTeaModel) Update(tea.Msg) (tea.Model, tea.Cmd)        { return fakeTeaModel{}, nil }
func (fakeTeaModel) View() string                               { return "" }

func TestHandleListActionShowReturnsSingleEnhancedError(t *testing.T) {
	profile.SetProfileDirOverride(t.TempDir())
	t.Cleanup(func() { profile.SetProfileDirOverride("") })

	// Ensure we do not hit the first-run branch for missing profiles.
	if _, err := profile.Save(&profile.Profile{Name: "default", CreatedAt: time.Now()}); err != nil {
		t.Fatalf("Save(default): %v", err)
	}

	m := tui.NewListModel([]*profile.Profile{{Name: "missing", CreatedAt: time.Now()}})
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result := updated.(tui.ListModel)

	err := handleListAction(result)
	if err == nil {
		t.Fatal("handleListAction() error = nil, want enhanced missing-profile error")
	}

	msg := err.Error()
	if !strings.Contains(msg, "profile 'missing' not found") {
		t.Fatalf("handleListAction() error = %q, want missing-profile message", msg)
	}
	if strings.Count(msg, "Use 'skel list' to see available profiles") != 1 {
		t.Fatalf("handleListAction() error = %q, want one guidance hint", msg)
	}
}

func TestHandleListActionShowSuggestionAppearsOnce(t *testing.T) {
	profile.SetProfileDirOverride(t.TempDir())
	t.Cleanup(func() { profile.SetProfileDirOverride("") })

	if _, err := profile.Save(&profile.Profile{Name: "default", CreatedAt: time.Now()}); err != nil {
		t.Fatalf("Save(default): %v", err)
	}

	m := tui.NewListModel([]*profile.Profile{{Name: "defalt", CreatedAt: time.Now()}})
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result := updated.(tui.ListModel)

	err := handleListAction(result)
	if err == nil {
		t.Fatal("handleListAction() error = nil, want enhanced suggested-profile error")
	}

	msg := err.Error()
	if !strings.Contains(msg, "profile 'defalt' not found") {
		t.Fatalf("handleListAction() error = %q, want missing-profile message", msg)
	}
	if strings.Count(msg, "Did you mean 'default'?") != 1 {
		t.Fatalf("handleListAction() error = %q, want one fuzzy suggestion", msg)
	}
	if strings.Count(msg, "Use 'skel list' to see all profiles") != 1 {
		t.Fatalf("handleListAction() error = %q, want one suggestion guidance line", msg)
	}
}

func TestRunListInteractiveShowSuggestionAppearsOnce(t *testing.T) {
	profile.SetProfileDirOverride(t.TempDir())
	t.Cleanup(func() { profile.SetProfileDirOverride("") })

	if _, err := profile.Save(&profile.Profile{Name: "default", CreatedAt: time.Now()}); err != nil {
		t.Fatalf("Save(default): %v", err)
	}

	origRunListProgram := runListProgram
	t.Cleanup(func() { runListProgram = origRunListProgram })
	runListProgram = func(_ tui.ListModel) (tea.Model, error) {
		model := tui.NewListModel([]*profile.Profile{{Name: "defalt", CreatedAt: time.Now()}})
		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
		return updated, nil
	}

	err := runListInteractive([]*profile.Profile{{Name: "defalt", CreatedAt: time.Now()}})
	if err == nil {
		t.Fatal("runListInteractive() error = nil, want enhanced suggested-profile error")
	}

	msg := err.Error()
	if strings.Count(msg, "Did you mean 'default'?") != 1 {
		t.Fatalf("runListInteractive() error = %q, want one fuzzy suggestion", msg)
	}
	if strings.Count(msg, "Use 'skel list' to see all profiles") != 1 {
		t.Fatalf("runListInteractive() error = %q, want one suggestion guidance line", msg)
	}
}

func TestRunListInteractiveEnhancesUnexpectedModelType(t *testing.T) {
	origRunListProgram := runListProgram
	t.Cleanup(func() { runListProgram = origRunListProgram })
	runListProgram = func(_ tui.ListModel) (tea.Model, error) {
		return fakeTeaModel{}, nil
	}

	err := runListInteractive([]*profile.Profile{{Name: "default", CreatedAt: time.Now()}})
	if err == nil {
		t.Fatal("runListInteractive() error = nil, want model type error")
	}

	if !strings.Contains(err.Error(), "unexpected model type from list") {
		t.Fatalf("runListInteractive() error = %q, want unexpected model type message", err.Error())
	}
}

