package tui

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/models/feed"
)

type feedListModel struct {
	repos feed.FeedRepository
	list  list.Model
	err   error
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
		if msg.String() == "enter" {
			selectedItem, ok := m.list.SelectedItem().(feedItem)
			if ok {
				slog.Debug("Feed selected", "id", selectedItem.id, "title", selectedItem.title)
				// Send a message to the main model to switch view
				return m, func() tea.Msg { return selectFeedMsg{feed: selectedItem} }
			}
		}
	}

	// Delegate other message processing to the list component
	// Delegate other message processing to the list component
	m.list, cmd = m.list.Update(msg)

	// Return the updated model (m) which is implicitly a tea.Model now
	return m, cmd
}

func (m feedListModel) View() string {
	slog.Debug("FeedList model View called")
	if m.err != nil {
		// Basic error display for now
		return fmt.Sprintf("Error fetching feeds: %v\n\nPress q to quit.", m.err)
	}
	// Render the list using its View method, wrapped in a basic style
	return docStyle.Render(m.list.View())
}
