package tui

import (
	"context"
	"fmt"
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/models/article"
	"github.com/mopemope/quicknews/tui/components"
	"github.com/toqueteos/webbrowser"
)

// Message indicating a feed has been selected
type selectFeedMsg struct {
	feed feedItem
}

// Message indicating an article has been selected
type selectArticleMsg struct {
	article articleItem
}

// Message to wrap fetched article content for summary view
type fetchedArticleContentMsg struct {
	article *ent.Article
	err     error
}

type model struct {
	client        *ent.Client
	articleRepos  article.ArticleRepository // Add article repository
	feedList      feedListModel
	articleList   articleListModel
	summaryView   summaryViewModel // Add summary view model
	currentView   viewState
	confirmDialog *components.ConfirmationDialog
	err           error
	windowWidth   int
	windowHeight  int
	config        *config.Config
}

type viewState int

const (
	feedListView viewState = iota
	articleListView
	summaryView // Add summary view state
)

func InitialModel(client *ent.Client, config *config.Config) model {
	return model{
		client:        client,
		articleRepos:  article.NewRepository(client), // Initialize article repository
		feedList:      newFeedListModel(client),
		articleList:   newArticleListModel(client, config),
		summaryView:   newSummaryViewModel(client, config), // Initialize summary view model
		currentView:   feedListView,
		confirmDialog: components.NewConfirmationDialog(),
		config:        config,
	}
}

func (m model) Init() tea.Cmd {
	slog.Debug("Main model Init called")
	// Initialize the view that is currently active
	switch m.currentView {
	case feedListView:
		return m.feedList.Init()
	case articleListView:
		return m.articleList.Init()
	case summaryView:
		return m.summaryView.Init()
	default:
		return nil // Add missing return
	}
}

// fetchArticleContentCmd fetches the full content of a single article for the summary view.
func (m *model) fetchArticleContentCmd(articleID uuid.UUID) tea.Cmd {
	slog.Debug("Fetching article content for summary", "articleID", articleID)
	ctx := context.Background()
	return func() tea.Msg {
		articleData, err := m.articleRepos.GetById(ctx, articleID) // Use main model's repo
		if err != nil {
			slog.Error("Failed to fetch article content for summary", "error", err, "articleID", articleID)
			// Return the error directly, the Update function will handle it
			return fmt.Errorf("failed to fetch article %s for summary: %w", articleID, err)
		}
		slog.Debug("Fetched article content successfully for summary", "articleID", articleID)
		return fetchedArticleContentMsg{article: articleData, err: nil}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	slog.Debug("Main model Update called", "msg", msg, "currentView", m.currentView)
	var cmd tea.Cmd

	// Handle different types of messages using helper methods
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle global key messages
		return m.handleGlobalKeyMsg(msg)
	case tea.WindowSizeMsg:
		// Handle window size changes
		m, cmd = m.handleWindowSizeMsg(msg)
		if cmd != nil {
			return m, cmd
		}
	case error:
		// Handle errors, potentially from fetch commands or sub-models
		m.err = msg
		slog.Error("Error received in main model", "error", msg, "currentView", m.currentView)
		return m, nil
	default:
		// Handle view-specific messages (messages that change view state)
		originalM := m
		m, cmd = m.handleViewSpecificMsg(msg)
		if cmd != nil {
			// If a view-specific handler returned a command, it means the view changed
			// so we should return early
			return m, cmd
		}
		// If no command was returned, reset m to the original state
		// to handle delegation with the unchanged model
		m = originalM
	}

	// If we reach this point, delegate updates to the appropriate subview
	m, cmd = m.handleViewDelegation(msg)
	return m, cmd
}

func (m model) View() string {
	slog.Debug("Main model View called", "currentView", m.currentView)
	if m.err != nil {
		// Prioritize showing error over delete confirmation? Or show both?
		// For now, error takes precedence.
		return "Error: " + m.err.Error()
	}

	// Render the current view
	currentViewContent := ""
	switch m.currentView {
	case feedListView:
		currentViewContent = m.feedList.View()
	case articleListView:
		return m.articleList.View()
	case summaryView:
		currentViewContent = m.summaryView.View()
	default:
		slog.Error("Unknown view state in View()", "viewState", m.currentView)
		currentViewContent = "Error: Unknown application state." // Provide a user-friendly error
	}

	// Otherwise, just return the current view's content
	return currentViewContent
}

func OpenArticleURL(url string) error {
	if err := webbrowser.Open(url); err != nil {
		return errors.Wrap(err, "failed to open URL in browser")
	}

	return nil
}
