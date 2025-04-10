package tui

import (
	"context"
	"fmt"
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/models/article"
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
	client               *ent.Client
	articleRepos         article.ArticleRepository // Add article repository
	feedList             feedListModel
	articleList          articleListModel
	summaryView          summaryViewModel // Add summary view model
	currentView          viewState
	showingDeleteConfirm bool // Flag to indicate if delete confirmation dialog is shown
	feedToDeleteTitle    string
	err                  error
	windowWidth          int
	windowHeight         int
	confirm              bool
	config               *config.Config
}

type viewState int

const (
	feedListView viewState = iota
	articleListView
	summaryView // Add summary view state
)

func InitialModel(client *ent.Client, config *config.Config, confirm bool) model {
	return model{
		client:       client,
		articleRepos: article.NewArticleRepository(client), // Initialize article repository
		feedList:     newFeedListModel(client),
		articleList:  newArticleListModel(client, confirm),
		summaryView:  newSummaryViewModel(client, config, confirm), // Initialize summary view model
		currentView:  feedListView,
		confirm:      confirm,
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
	slog.Debug("Main model Update called", "msg", msg, "currentView", m.currentView, "showingDeleteConfirm", m.showingDeleteConfirm)
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global keybindings when not in delete confirm dialog
		switch msg.String() {
		case "ctrl+c", "q":
			// If in delete confirm dialog, 'q' should probably cancel it first.
			// However, the logic above handles 'esc', which is common for cancelling.
			// If 'q' should also cancel, add it to the `if m.showingDeleteConfirm` block.
			// If 'q' should always quit, this placement is fine.
			return m, tea.Quit
		}
		// Allow key presses to fall through to the current view's Update if not handled globally
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		// Removed duplicate assignment: m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		slog.Debug("Window size changed", "width", m.windowWidth, "height", m.windowHeight)
		// Propagate window size to all views, handling both return values and type assertion
		// Rename variable to avoid conflict with type name
		var updatedFeedListModel tea.Model
		updatedFeedListModel, _ = m.feedList.Update(msg) // Ignore command for window size propagation
		// Use the renamed variable in the type assertion
		if flm, ok := updatedFeedListModel.(feedListModel); ok {
			m.feedList = flm
		} else {
			slog.Error("Update returned unexpected type for feedListModel during window resize")
			// Optionally handle the error, e.g., set m.err
		}
		m.articleList, _ = m.articleList.Update(msg) // No cmd expected here usually

	case selectFeedMsg: // Handle feed selection
		m.currentView = articleListView
		// Pass current window dimensions to SetFeedID for initial layout
		cmd = m.articleList.SetFeed(msg.feed, m.windowWidth, m.windowHeight)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...) // Return early as view changed

	case selectArticleMsg: // Handle article selection from article list
		slog.Debug("Received selectArticleMsg", "articleTitle", msg.article.title)
		m.currentView = summaryView
		m.err = nil // Clear previous errors
		// Fetch full article content
		cmd = m.fetchArticleContentCmd(msg.article.id)
		cmds = append(cmds, cmd)
		// Don't set content yet, wait for fetchedArticleContentMsg
		return m, tea.Batch(cmds...)

	case fetchedArticleContentMsg: // Handle fetched article content for summary view
		if msg.err != nil {
			slog.Error("Error fetching article content for summary", "error", msg.err)
			m.err = msg.err // Show error in main model? Or delegate to summary view?
			// Potentially switch back to article list on error?
			// m.currentView = articleListView
		} else {
			slog.Debug("Received fetched article content for summary", "articleID", msg.article.ID)
			m.err = nil
			// Set content in summary view
			cmd = m.summaryView.SetContent(msg.article, m.windowWidth, m.windowHeight)
			cmds = append(cmds, cmd)
		}
		// Ensure we are in the summary view even if fetch failed, error will be shown
		m.currentView = summaryView
		return m, tea.Batch(cmds...)

	case backToFeedListMsg: // Handle going back from article list to feed list
		slog.Debug("Received backToFeedListMsg")
		m.currentView = feedListView
		m.err = nil // Clear any errors from the previous view
		// Optionally, trigger a refresh or Init for the feed list if needed
		// cmd = m.feedList.Init() // Example: Re-initialize feed list
		// cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...) // Return early as view changed

	case backToArticleListMsg: // Handle going back from summary view to article list
		slog.Debug("Received backToArticleListMsg")
		m.currentView = articleListView
		m.err = nil // Clear any errors from the summary view
		// Trigger article list refetch
		cmd = m.articleList.fetchArticlesCmd()
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...) // Return early as view changed

	case error:
		// Handle errors, potentially from fetch commands or sub-models
		// Clear delete confirmation state if an error occurs elsewhere
		m.showingDeleteConfirm = false
		m.err = msg
		slog.Error("Error received in main model", "error", msg, "currentView", m.currentView)
		// Decide if the error should be displayed globally or handled by the current view
		// For now, let the current view handle its own errors primarily.
		// return m, nil
	}

	// Delegate updates to the current view's model
	switch m.currentView {
	case feedListView:
		// Update returns tea.Model, so we need a type assertion
		var updatedModel tea.Model
		updatedModel, cmd = m.feedList.Update(msg)
		if flm, ok := updatedModel.(feedListModel); ok {
			m.feedList = flm
		} else {
			slog.Error("Update returned unexpected type for feedListModel")
			// Optionally handle the error, e.g., set m.err
		}
		cmds = append(cmds, cmd)
	case articleListView:
		var updatedArticleList articleListModel // Keep this if articleListModel.Update returns articleListModel
		// Remove redeclaration: var updatedArticleList articleListModel
		updatedArticleList, cmd = m.articleList.Update(msg)
		m.articleList = updatedArticleList
		cmds = append(cmds, cmd)
	case summaryView:
		var updatedSummaryView summaryViewModel
		updatedSummaryView, cmd = m.summaryView.Update(msg)
		m.summaryView = updatedSummaryView
		cmds = append(cmds, cmd)
	default:
		slog.Warn("Unhandled view state in main update", "viewState", m.currentView)
	}

	// If the main model stored an error, clear it if the sub-model handled an update successfully?
	// This logic depends on how errors are intended to be surfaced.
	// if m.err != nil && cmd == nil { // Example: Clear main error if sub-update occurred without new cmds/errors
	//  m.err = nil
	// }

	return m, tea.Batch(cmds...)
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

	// If showing delete confirmation, render it on top
	if m.showingDeleteConfirm {
		dialogText := fmt.Sprintf("\n\n   Delete feed '%s'? (y/N) \n\n", m.feedToDeleteTitle)
		// Basic dialog styling (can be improved with lipgloss)
		dialogBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")). // Purple border
			Padding(1, 2).
			Render(dialogText)

		// Center the dialog (approximately)
		// dialogWidth := lipgloss.Width(dialogBox)   // Keep if needed for Place logic adjustments
		// dialogHeight := lipgloss.Height(dialogBox) // Keep if needed for Place logic adjustments
		// x and y are calculated but not used when using lipgloss.Center positioning
		// x := (m.windowWidth - dialogWidth) / 2
		// y := (m.windowHeight - dialogHeight) / 2

		// Use lipgloss.Place to position the dialog over the current view content
		// Note: This requires knowing the full window dimensions.
		// We might need a more robust way to overlay if views don't occupy the full screen.
		// For now, let's assume the current view takes the full screen.
		return lipgloss.Place(m.windowWidth, m.windowHeight,
			lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center, currentViewContent, dialogBox), // Attempt to stack, might need adjustment
			lipgloss.WithWhitespaceChars(" "),                                     // Fill remaining space
			lipgloss.WithWhitespaceForeground(lipgloss.AdaptiveColor{Light: "#FFF", Dark: "#000"}),
		)

		// Simpler approach: Just render the dialog after the main content (less visually appealing overlay)
		// return currentViewContent + "\n" + dialogBox
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
