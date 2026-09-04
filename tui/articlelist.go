package tui

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/models/article"
	"github.com/mopemope/quicknews/models/feed"
	"github.com/mopemope/quicknews/models/summary"
	"github.com/mopemope/quicknews/tui/components"
)

// Message to indicate going back to the feed list
type backToFeedListMsg struct{}

// articleListRefreshMsg triggers a re-fetch of the article list.
// Returned from confirmation dialog actions where the closure captures a
// value copy of the model, so list mutations must flow through Update.
type articleListRefreshMsg struct{}

type articleListModel struct {
	feedRepos     feed.FeedRepository
	repos         article.ArticleRepository
	summaryRepos  summary.SummaryRepository
	ctx           context.Context
	list          list.Model
	feed          feedItem
	listWidth     int
	showAll       bool
	err           error
	confirmDialog *components.ConfirmationDialog
	config        *config.Config
}

type articleItem struct {
	id           uuid.UUID
	title        string
	publishedAt  *time.Time
	link         string
	summaryTitle string
	summaryCount int
	isBookmark   bool
	isRead       bool
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
	if i.isRead {
		title = "✓ " + title
	}
	if i.publishedAt != nil {
		title = fmt.Sprintf("%s (%s)", title, i.publishedAt.Local().Format("2006-01-02 15:04"))
		stitle = fmt.Sprintf("%s [%d] (%s)", stitle, i.summaryCount, i.publishedAt.Local().Format("2006-01-02 15:04"))
	}
	return fmt.Sprintf("%s\n%s", stitle, title)
}

func (i articleItem) Description() string { return i.link }

func (i articleItem) FilterValue() string { return i.title }

func newArticleListModel(ctx context.Context, client *ent.Client, config *config.Config) articleListModel {
	if ctx == nil {
		ctx = context.Background()
	}

	defaultDelegate := list.NewDefaultDelegate()

	l := list.New([]list.Item{}, defaultDelegate, 0, 0)
	l.Title = "Articles"

	return articleListModel{
		feedRepos:     feed.NewRepository(client),
		repos:         article.NewRepository(client),
		summaryRepos:  summary.NewRepository(client),
		ctx:           ctx,
		list:          l,
		confirmDialog: components.NewConfirmationDialog(),
		config:        config,
	}
}

// SetFeed sets the feed for which to fetch articles, updates layout, and triggers fetching.
func (m *articleListModel) SetFeed(feed feedItem, width, height int) tea.Cmd {
	m.feed = feed
	// m.selectedArticle = nil // Removed
	m.list.Title = "Articles"      // Reset title potentially
	m.list.SetItems([]list.Item{}) // Clear previous items
	m.err = nil
	m.showAll = false // Reset to unread-only view when switching feeds
	m.updateListTitle()
	// Update list size immediately when setting feed

	slog.Debug("ArticleList SetFeed called", "width", width, "height", height, "listHeight", m.list.Height())
	return m.fetchArticlesCmd()
}

func (m *articleListModel) updateListTitle() {
	if m.showAll {
		m.list.Title = fmt.Sprintf("Articles - %s (all)", m.feed.title)
	} else {
		m.list.Title = fmt.Sprintf("Articles - %s", m.feed.title)
	}
}

// fetchArticlesCmd fetches articles for the current feedID from the database.
func (m *articleListModel) fetchArticlesCmd() tea.Cmd {

	return func() tea.Msg {
		var articles ent.Articles
		var err error
		if m.showAll {
			articles, err = m.repos.GetByFeed(m.ctx, m.feed.id)
		} else {
			articles, err = m.repos.GetByUnreaded(m.ctx, m.feed.id)
		}
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
			isRead := false
			if a.Edges.Summary != nil {
				summaryTitle = a.Edges.Summary.Title
				count = len([]rune(a.Edges.Summary.Summary))
				isRead = a.Edges.Summary.Readed
			}
			items[i] = articleItem{
				id:           a.ID,
				title:        a.Title,
				publishedAt:  publishedAtPtr, // Pass the pointer
				link:         a.URL,
				summaryTitle: summaryTitle,
				summaryCount: count,
				isBookmark:   a.Edges.Feed != nil && a.Edges.Feed.IsBookmark,
				isRead:       isRead,
			}
		}
		return items // Return fetched items as message
	}
}

// refreshItemReadStatus updates the read marker of the item with the given article ID in place.
func (m *articleListModel) refreshItemReadStatus(id uuid.UUID, isRead bool) {
	items := m.list.Items()
	for i, item := range items {
		ai, ok := item.(articleItem)
		if !ok || ai.id != id {
			continue
		}
		ai.isRead = isRead
		items[i] = ai
		m.list.SetItems(items)
		break
	}
}

// applyReadResult updates the list after toggling the read status of an article.
// In unread-only mode the item is removed; in show-all mode the read marker is refreshed.
func (m *articleListModel) applyReadResult(id uuid.UUID, isRead bool) {
	if !m.showAll {
		m.list.RemoveItem(m.list.Index())
		return
	}
	m.refreshItemReadStatus(id, isRead)
}

func (m articleListModel) Init() tea.Cmd {
	slog.Debug("ArticleList model Init called")
	// Initial fetching is triggered by SetFeed
	return nil
}

func (m articleListModel) Update(msg tea.Msg) (articleListModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// Handle confirmation dialog if active
	if handled, dialogCmd := m.confirmDialog.Update(msg); handled {
		return m, dialogCmd
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

	case articleListRefreshMsg: // Re-fetch articles after a confirm dialog action
		return m, m.fetchArticlesCmd()

	case tea.KeyMsg:

		switch msg.String() {
		case "b": // Go back to feed list view
			slog.Debug("Back key pressed in article list")
			return m, func() tea.Msg { return backToFeedListMsg{} } // Send message to main model
		case "r": // Reload articles
			slog.Debug("Reloading articles")
			cmds = append(cmds, m.fetchArticlesCmd()) // Trigger article fetch
		case "a": // Toggle between unread-only and all articles
			slog.Debug("Toggling show all articles", "showAll", !m.showAll)
			m.showAll = !m.showAll
			m.updateListTitle()
			cmds = append(cmds, m.fetchArticlesCmd())
		case "o":
			selectedItem, ok := m.list.SelectedItem().(articleItem)
			if ok {
				if err := OpenArticleURL(selectedItem.link); err != nil {
					slog.Error("Failed to open url", "error", err)
				}
			}
		case "R": // Toggle read status of the selected article
			selectedItem, ok := m.list.SelectedItem().(articleItem)
			// bookmark is not allowed to be readed
			if ok && !selectedItem.isBookmark {
				id := selectedItem.id
				article, err := m.repos.GetById(m.ctx, id)
				if err != nil {
					slog.Error("Failed to get article by ID", "error", err)
					return m, nil
				}
				if article.Edges.Summary == nil {
					slog.Error("Summary edge not loaded for article", "articleID", id)
					return m, nil
				}

				newReaded := !selectedItem.isRead
				summaryID := article.Edges.Summary.ID
				if newReaded {
					if m.config.RequireConfirm {
						m.confirmDialog.Show(
							"記事を既読にしますか？ (y/N)",
							func() tea.Cmd {
								return func() tea.Msg {
									if err := m.summaryRepos.SetReaded(m.ctx, summaryID, true); err != nil {
										slog.Error("Failed to mark as read", "error", err)
										return errors.Wrap(err, "failed to mark article as read")
									}
									// Re-fetch via message so the current model updates its list.
									return articleListRefreshMsg{}
								}
							},
							nil,
						)
					} else {
						if err := m.summaryRepos.SetReaded(m.ctx, summaryID, true); err != nil {
							slog.Error("Failed to mark as read", "error", err)
							return m, nil
						}
						m.applyReadResult(id, true)
					}
				} else {
					if m.config.RequireConfirm {
						m.confirmDialog.Show(
							"記事を未読に戻しますか？ (y/N)",
							func() tea.Cmd {
								return func() tea.Msg {
									if err := m.summaryRepos.SetReaded(m.ctx, summaryID, false); err != nil {
										slog.Error("Failed to mark as unread", "error", err)
										return errors.Wrap(err, "failed to mark article as unread")
									}
									// Re-fetch via message so the current model updates its list.
									return articleListRefreshMsg{}
								}
							},
							nil,
						)
					} else {
						if err := m.summaryRepos.SetReaded(m.ctx, summaryID, false); err != nil {
							slog.Error("Failed to mark as unread", "error", err)
							return m, nil
						}
						m.applyReadResult(id, false)
					}
				}
				return m, nil
			}
		case "A": // Mark all articles in the current feed as read
			// bookmark is not allowed to be readed
			if !m.feed.isBookmark {
				feedID := m.feed.id
				refetch := func() tea.Msg {
					if _, err := m.summaryRepos.MarkFeedReaded(m.ctx, feedID); err != nil {
						slog.Error("Failed to mark feed as read", "error", err)
						return errors.Wrap(err, "failed to mark feed as read")
					}
					// Re-fetch via message so the current model updates its list.
					return articleListRefreshMsg{}
				}
				if m.config.RequireConfirm {
					m.confirmDialog.Show(
						"このフィードの全記事を既読にしますか？ (y/N)",
						func() tea.Cmd {
							return refetch
						},
						nil,
					)
				} else {
					return m, refetch
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

	if m.confirmDialog.IsActive() {
		return m.confirmDialog.View(m.list.Width(), m.list.Height())
	}

	// 通常表示
	return content
}
