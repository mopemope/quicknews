package tui

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/models/article"
	"github.com/mopemope/quicknews/models/feed"
	"github.com/mopemope/quicknews/models/summary"
)

// Message to indicate going back to the feed list
type backToFeedListMsg struct{}

type articleListModel struct {
	feedRepos         feed.FeedRepository
	repos             article.ArticleRepository
	summaryRepos      summary.SummaryRepository
	list              list.Model
	feed              feedItem
	listWidth         int
	err               error
	showConfirmDialog bool
	confirmDialogMsg  string
	onConfirmYes      func() tea.Cmd
	onConfirmNo       func() tea.Cmd
	config            *config.Config
}

type articleItem struct {
	id           uuid.UUID
	title        string
	publishedAt  *time.Time
	link         string
	summaryTitle string
	summaryCount int
	isBookmark   bool
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
		stitle = fmt.Sprintf("%s [%d] (%s)", stitle, i.summaryCount, i.publishedAt.Local().Format("2006-01-02 15:04"))
	}
	return fmt.Sprintf("%s\n%s", stitle, title)
}

func (i articleItem) Description() string { return i.link }

func (i articleItem) FilterValue() string { return i.title }

func newArticleListModel(client *ent.Client, config *config.Config) articleListModel {
	defaultDelegate := list.NewDefaultDelegate()

	l := list.New([]list.Item{}, defaultDelegate, 0, 0)
	l.Title = "Articles"

	return articleListModel{
		feedRepos:         feed.NewFeedRepository(client),
		repos:             article.NewArticleRepository(client),
		summaryRepos:      summary.NewSummaryRepository(client),
		list:              l,
		showConfirmDialog: false,
		config:            config,
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
		articles, err := m.repos.GetByUnreaded(ctx, m.feed.id)
		if err != nil {
			slog.Error("Failed to fetch articles", "error", err, "feedID", m.feed.id)
			return errors.Wrapf(err, "failed to fetch articles for feed %s: %w", m.feed.id)
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
			summaryTitle := a.Title
			count := 0
			if a.Edges.Summary != nil {
				summaryTitle = a.Edges.Summary.Title
				count = len([]rune(a.Edges.Summary.Summary))
			}
			items[i] = articleItem{
				id:           a.ID,
				title:        a.Title,
				publishedAt:  publishedAtPtr, // Pass the pointer
				link:         a.URL,
				summaryTitle: summaryTitle,
				summaryCount: count,
				isBookmark:   a.Edges.Feed != nil && a.Edges.Feed.IsBookmark,
			}
		}
		return items // Return fetched items as message
	}
}

func (m articleListModel) Init() tea.Cmd {
	slog.Debug("ArticleList model Init called")
	// Initial fetching is triggered by SetFeed
	return nil
}

// ShowConfirmationDialog shows a confirmation dialog with the provided message
// and executes the provided callbacks when Yes or No is selected
func (m *articleListModel) ShowConfirmationDialog(message string, onYes, onNo func() tea.Cmd) {
	m.showConfirmDialog = true
	m.confirmDialogMsg = message
	m.onConfirmYes = onYes
	m.onConfirmNo = onNo
}

// HideConfirmationDialog hides the confirmation dialog
func (m *articleListModel) HideConfirmationDialog() {
	m.showConfirmDialog = false
	m.confirmDialogMsg = ""
	m.onConfirmYes = nil
	m.onConfirmNo = nil
}

func (m articleListModel) Update(msg tea.Msg) (articleListModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// Handle confirmation dialog first if it's active
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
		case "o":
			selectedItem, ok := m.list.SelectedItem().(articleItem)
			if ok {
				if err := OpenArticleURL(selectedItem.link); err != nil {
					slog.Error("Failed to open url", "error", err)
				}
			}
		case "R":
			selectedItem, ok := m.list.SelectedItem().(articleItem)
			// bookmark is not allowed to be readed
			if ok && !selectedItem.isBookmark {
				id := selectedItem.id
				ctx := context.Background()
				article, err := m.repos.GetById(ctx, id)
				if err != nil {
					slog.Error("Failed to get article by ID", "error", err)
					return m, nil
				}

				if m.config.RequireConfirm {
					m.ShowConfirmationDialog(
						"記事を既読にしますか？ (y/N)",
						func() tea.Cmd {
							ctx := context.Background()
							return func() tea.Msg {
								if err := m.summaryRepos.UpdateReaded(ctx, article.Edges.Summary); err != nil {
									slog.Error("Failed to mark as read", "error", err)
									return errors.Wrap(err, "failed to mark article as read")
								}
								m.list.RemoveItem(m.list.Index())
								return nil
							}
						},
						nil,
					)
				} else {
					if err := m.summaryRepos.UpdateReaded(ctx, article.Edges.Summary); err != nil {
						slog.Error("Failed to mark as read", "error", err)
						return m, nil
					}
					m.list.RemoveItem(m.list.Index())
				}
				return m, nil
			}
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
		}
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)     // Append the command from the list update
	return m, tea.Batch(cmds...) // Return updated model and commands
}

func (m articleListModel) View() string {
	slog.Debug("ArticleList model View called", "listWidth", m.listWidth)

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

	// 通常表示
	return content
}
