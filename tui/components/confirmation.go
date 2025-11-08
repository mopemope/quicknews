package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmationDialog represents a reusable confirmation dialog component
type ConfirmationDialog struct {
	active       bool
	message      string
	onConfirmYes func() tea.Cmd
	onConfirmNo  func() tea.Cmd
	dialogWidth  int
	dialogStyle  lipgloss.Style
}

// NewConfirmationDialog creates a new confirmation dialog component
func NewConfirmationDialog() *ConfirmationDialog {
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		Width(40)

	return &ConfirmationDialog{
		dialogStyle: dialogStyle,
		dialogWidth: 40,
	}
}

// IsActive returns true if the confirmation dialog is currently active
func (cd *ConfirmationDialog) IsActive() bool {
	return cd.active
}

// Show activates the confirmation dialog
func (cd *ConfirmationDialog) Show(message string, onYes, onNo func() tea.Cmd) {
	cd.active = true
	cd.message = message
	cd.onConfirmYes = onYes
	cd.onConfirmNo = onNo
}

// Hide deactivates the confirmation dialog
func (cd *ConfirmationDialog) Hide() {
	cd.active = false
	cd.message = ""
	cd.onConfirmYes = nil
	cd.onConfirmNo = nil
}

// Update handles key events for the confirmation dialog
func (cd *ConfirmationDialog) Update(msg tea.Msg) (bool, tea.Cmd) {
	if !cd.active {
		return false, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			cd.active = false
			if cd.onConfirmYes != nil {
				return true, cd.onConfirmYes()
			}
			return true, nil
		case "n", "N", "esc":
			cd.active = false
			if cd.onConfirmNo != nil {
				return true, cd.onConfirmNo()
			}
			return true, nil
		default:
			// Ignore other keypresses when dialog is shown
			return true, nil
		}
	}

	return false, nil
}

// View returns the rendered view of the confirmation dialog
func (cd *ConfirmationDialog) View(width, height int) string {
	if !cd.active {
		return ""
	}

	dialogBox := cd.dialogStyle.Render(cd.message)

	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		dialogBox,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
	)
}
