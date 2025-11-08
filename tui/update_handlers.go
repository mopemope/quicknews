package tui

import (
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
)

// handleGlobalKeyMsg handles global key messages that apply regardless of current view
func (m model) handleGlobalKeyMsg(msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	}
	// Allow key presses to fall through to the current view's Update if not handled globally
	return m, nil
}

// handleWindowSizeMsg handles window size changes and propagates to subviews
func (m model) handleWindowSizeMsg(msg tea.WindowSizeMsg) (model, tea.Cmd) {
	m.windowWidth = msg.Width
	m.windowHeight = msg.Height
	// Propagate window size to all views, handling both return values and type assertion
	var updatedFeedListModel tea.Model
	updatedFeedListModel, _ = m.feedList.Update(msg) // Ignore command for window size propagation
	// Use the renamed variable in the type assertion
	if flm, ok := updatedFeedListModel.(feedListModel); ok {
		m.feedList = flm
	} else {
		slog.Error("Update returned unexpected type for feedListModel during window resize")
	}
	m.articleList, _ = m.articleList.Update(msg) // No cmd expected here usually
	return m, nil
}

// handleViewSpecificMsg handles messages that change the view state
func (m model) handleViewSpecificMsg(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
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
	}

	return m, nil
}

// handleViewDelegation handles delegating updates to the appropriate subview
func (m model) handleViewDelegation(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

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
		}
		cmds = append(cmds, cmd)
	case articleListView:
		var updatedArticleList articleListModel // Keep this if articleListModel.Update returns articleListModel
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

	return m, tea.Batch(cmds...)
}
