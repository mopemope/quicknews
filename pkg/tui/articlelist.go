package tui

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/models/article"
	"github.com/mopemope/quicknews/models/feed"
)

// Message to indicate going back to the feed list
type backToFeedListMsg struct{}

type articleListModel struct {
	feedRepos feed.FeedRepository
	repos     article.ArticleRepository
	list      list.Model
	feed      feedItem
	// selectedArticle *ent.Article // Removed: Selection handled by main model
	listWidth int
	err       error
}

// Removed fetchedArticleContentMsg, handled by main model

type articleItem struct {
	id           uuid.UUID
	title        string
	publishedAt  *time.Time
	link         string
	summaryTitle string
}

func (i articleItem) Title() string {
	title := i.title
	stitle := i.summaryTitle
	if title == "" {
		title = "No title"
	}
	if stitle == "" {
		stitle = "No title"
	}
	if i.publishedAt != nil {
		title = fmt.Sprintf("%s (%s)", title, i.publishedAt.Local().Format("2006-01-02 15:04"))
		stitle = fmt.Sprintf("%s (%s)", stitle, i.publishedAt.Local().Format("2006-01-02 15:04"))
	}
	return fmt.Sprintf("%s\n%s", stitle, title)
}

func (i articleItem) Description() string { return i.link }

func (i articleItem) FilterValue() string { return i.title }

func newArticleListModel(client *ent.Client) articleListModel {
	defaultDelegate := list.NewDefaultDelegate()

	l := list.New([]list.Item{}, defaultDelegate, 0, 0)
	l.Title = "Articles"

	return articleListModel{
		feedRepos: feed.NewFeedRepository(client),
		repos:     article.NewArticleRepository(client),
		list:      l,
	}
}

// SetFeed sets the feed for which to fetch articles, updates layout, and triggers fetching.
func (m *articleListModel) SetFeed(feed feedItem, width, height int) tea.Cmd {
	m.feed = feed
	// m.selectedArticle = nil // Removed
	m.list.Title = "Articles"      // Reset title potentially
	m.list.SetItems([]list.Item{}) // Clear previous items
	m.err = nil
	m.list.Title = fmt.Sprintf("Articles - %s", feed.title)
	// Update list size immediately when setting feed

	slog.Debug("ArticleList SetFeed called", "width", width, "height", height, "listHeight", m.list.Height())
	return m.fetchArticlesCmd()
}

// fetchArticlesCmd fetches articles for the current feedID from the database.
func (m *articleListModel) fetchArticlesCmd() tea.Cmd {

	ctx := context.Background()
	return func() tea.Msg {
		articles, err := m.repos.GetByFeed(ctx, m.feed.id)
		if err != nil {
			slog.Error("Failed to fetch articles", "error", err, "feedID", m.feed.id)
			return fmt.Errorf("failed to fetch articles for feed %s: %w", m.feed.id, err)
		}
		slog.Debug("Fetched articles successfully", "count", len(articles), "feedID", m.feed.id)

		items := make([]list.Item, len(articles))
		for i, a := range articles {
			// Assign the address of a.PublishedAt if it's not the zero value,
			// otherwise keep it nil. Check if PublishedAt is nullable or handle zero time.
			// For now, directly assign the address assuming PublishedAt is always set.
			var publishedAtPtr *time.Time
			if !a.PublishedAt.IsZero() {
				publishedAtPtr = &a.PublishedAt
			}
			items[i] = articleItem{
				id:           a.ID,
				title:        a.Title,
				publishedAt:  publishedAtPtr, // Pass the pointer
				link:         a.URL,
				summaryTitle: a.Edges.Summary.Title,
			}
		}
		return items // Return fetched items as message
	}
}

// Removed fetchArticleContentCmd, handled by main model

func (m articleListModel) Init() tea.Cmd {
	slog.Debug("ArticleList model Init called")
	// Initial fetching is triggered by SetFeed
	return nil
}

func (m articleListModel) Update(msg tea.Msg) (articleListModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height-15)
		slog.Info("ArticleList window size updated", "width", msg.Width, "height", msg.Height)

	case []list.Item: // Received fetched article items from fetchArticlesCmd
		slog.Info("Received fetched article items", "count", len(msg))
		m.list.SetItems(msg)
		m.err = nil // Clear previous errors

	case error: // General errors or errors from fetchArticlesCmd
		slog.Error("Error received in ArticleList model", "error", msg)
		m.err = msg
		// m.selectedArticle = nil // Removed
		return m, nil

	case tea.KeyMsg:

		switch msg.String() {
		case "b": // Go back to feed list view
			slog.Debug("Back key pressed in article list")
			return m, func() tea.Msg { return backToFeedListMsg{} } // Send message to main model
		case "enter":
			selectedItem, ok := m.list.SelectedItem().(articleItem)
			if ok {
				slog.Info("Enter key pressed, selecting article", "articleID", selectedItem.id, "title", selectedItem.title)
				// Send message to main model to handle selection
				cmd = func() tea.Msg {
					return selectArticleMsg{article: selectedItem}
				}
				cmds = append(cmds, cmd)
			}
			// Viewport scrolling is handled in the summary view.
			// For now, rely on list navigation triggering content load.
		}
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)     // Append the command from the list update
	return m, tea.Batch(cmds...) // Return updated model and commands
}

func (m articleListModel) View() string {
	slog.Debug("ArticleList model View called", "listWidth", m.listWidth)
	// Ensure the list view itself is rendered within the docStyle margin
	return docStyle.Render(m.list.View())
}
