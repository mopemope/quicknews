package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"testing"
)

func TestConfirmationDialog(t *testing.T) {
	dialog := NewConfirmationDialog()

	// Test initial state
	if dialog.IsActive() {
		t.Error("Expected dialog to not be active initially")
	}

	// Test showing the dialog
	called := false
	onYes := func() tea.Cmd {
		called = true
		return nil
	}

	dialog.Show("Test message", onYes, nil)

	if !dialog.IsActive() {
		t.Error("Expected dialog to be active after showing")
	}

	if dialog.message != "Test message" {
		t.Error("Expected message to be set")
	}

	// Test handling 'y' key
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	handled, _ := dialog.Update(msg)

	if !handled {
		t.Error("Expected dialog to handle the message")
	}

	if !called {
		t.Error("Expected onYes callback to be called")
	}

	if dialog.IsActive() {
		t.Error("Expected dialog to not be active after handling 'y'")
	}
}
