package tui

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/models/feed"
)

type feedListModel struct {
	repos             feed.FeedRepository
	list              list.Model
	err               error
	showConfirmDialog bool
	confirmDialogMsg  string
	onConfirmYes      func() tea.Cmd
	onConfirmNo       func() tea.Cmd
}

type feedItem struct {
	id    uuid.UUID
	title string
	url   string
}

func (i feedItem) Title() string       { return i.title }
func (i feedItem) Description() string { return i.url } // Show URL in description for now
func (i feedItem) FilterValue() string { return i.title }

func newFeedListModel(client *ent.Client) feedListModel {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Feeds"

	return feedListModel{
		repos: feed.NewFeedRepository(client),
		list:  l,
	}
}

// fetchFeedsCmd fetches feeds from the database.
func (m *feedListModel) fetchFeedsCmd() tea.Msg {
	slog.Debug("Fetching feeds from database")

	ctx := context.Background()
	feeds, err := m.repos.All(ctx)
	if err != nil {
		slog.Error("Failed to fetch feeds", "error", err)
		return errors.Wrap(err, "failed to fetch feeds")
	}

	slog.Debug("Fetched feeds successfully", "count", len(feeds))

	items := make([]list.Item, len(feeds))
	for i, f := range feeds {
		items[i] = feedItem{id: f.ID, title: f.Title, url: f.URL}
	}
	return items // Return fetched items as message
}

func (m feedListModel) Init() tea.Cmd {
	slog.Debug("FeedList model Init called")
	return m.fetchFeedsCmd // Use the command directly
}

// Update method now returns tea.Model to satisfy the interface
func (m feedListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	slog.Debug("FeedList model Update called", "msg", msg)
	var cmd tea.Cmd
	var cmds []tea.Cmd

	if m.showConfirmDialog {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "y", "Y":
				m.showConfirmDialog = false
				if m.onConfirmYes != nil {
					cmd = m.onConfirmYes()
					cmds = append(cmds, cmd)
				}
				return m, tea.Batch(cmds...)
			case "n", "N", "esc":
				m.showConfirmDialog = false
				if m.onConfirmNo != nil {
					cmd = m.onConfirmNo()
					cmds = append(cmds, cmd)
				}
				return m, tea.Batch(cmds...)
			default:
				// Ignore other keypresses when dialog is shown
				return m, nil
			}
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Adjust list size based on window dimensions
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		slog.Debug("FeedList window size updated", "width", msg.Width-h, "height", msg.Height-v)
	case []list.Item: // Received fetched feed items
		slog.Debug("Received fetched feed items", "count", len(msg))
		m.list.SetItems(msg)
		m.err = nil // Clear previous errors if fetch is successful
	case error: // Received an error (e.g., from fetchFeedsCmd)
		slog.Error("Error received in FeedList model", "error", msg)
		m.err = msg
		return m, nil // Stop processing further on error

	case tea.KeyMsg:
		// Handle Enter key press for feed selection
		switch msg.String() {
		case "enter":
			selectedItem, ok := m.list.SelectedItem().(feedItem)
			if ok {
				slog.Debug("Feed selected", "id", selectedItem.id, "title", selectedItem.title)
				// Send a message to the main model to switch view
				return m, func() tea.Msg { return selectFeedMsg{feed: selectedItem} }
			}
		case "d":
			selectedItem, ok := m.list.SelectedItem().(feedItem)
			if ok {
				m.ShowConfirmationDialog(
					"このフィード削除しますか？ (y/N)",
					func() tea.Cmd {
						ctx := context.Background()
						return func() tea.Msg {
							if err := m.repos.DeleteWithArticle(ctx, selectedItem.id); err != nil {
								slog.Error("Failed to mark as read", "error", err)
								return errors.Wrap(err, "failed to mark article as read")
							}
							m.list.RemoveItem(m.list.Index())
							return m.fetchFeedsCmd()
						}
					},
					nil,
				)
			}
		}
	}

	// Delegate other message processing to the list component
	m.list, cmd = m.list.Update(msg)

	// Return the updated model (m) which is implicitly a tea.Model now
	return m, cmd
}

func (m *feedListModel) ShowConfirmationDialog(message string, onYes, onNo func() tea.Cmd) {
	m.showConfirmDialog = true
	m.confirmDialogMsg = message
	m.onConfirmYes = onYes
	m.onConfirmNo = onNo
}

// HideConfirmationDialog hides the confirmation dialog
func (m *feedListModel) HideConfirmationDialog() {
	m.showConfirmDialog = false
	m.confirmDialogMsg = ""
	m.onConfirmYes = nil
	m.onConfirmNo = nil
}

func (m feedListModel) View() string {
	slog.Debug("FeedList model View called")
	if m.err != nil {
		// Basic error display for now
		return fmt.Sprintf("Error fetching feeds: %v\n\nPress q to quit.", m.err)
	}
	// Render the list using its View method, wrapped in a basic style
	content := docStyle.Render(m.list.View())
	if m.showConfirmDialog {
		dialogStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2).
			Width(40)

		dialogBox := dialogStyle.Render(m.confirmDialogMsg)

		return lipgloss.Place(
			m.list.Width(),
			m.list.Height(),
			lipgloss.Center,
			lipgloss.Center,
			dialogBox,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
		)
	}
	return content
}
