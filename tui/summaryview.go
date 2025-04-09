package tui

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cockroachdb/errors"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/models/article"
	"github.com/mopemope/quicknews/models/summary"
	"github.com/mopemope/quicknews/tts"
)

// Message to indicate going back to the article list
type backToArticleListMsg struct{}

type summaryViewModel struct {
	viewport          viewport.Model
	article           *ent.Article
	ready             bool // Indicates if the viewport is ready
	summaryRepos      summary.SummaryRepository
	articleRepos      article.ArticleRepository // Add ArticleRepository
	confirm           bool
	showConfirmDialog bool
	confirmDialogMsg  string
	onConfirmYes      func() tea.Cmd
	onConfirmNo       func() tea.Cmd
}

func newSummaryViewModel(client *ent.Client, confirm bool) summaryViewModel {
	vp := viewport.New(0, 0) // Initial size, will be updated
	vp.Style = summaryViewStyle
	return summaryViewModel{
		viewport:     vp,
		summaryRepos: summary.NewSummaryRepository(client),
		articleRepos: article.NewArticleRepository(client), // Initialize ArticleRepository
		confirm:      confirm,
	}
}

// SetContent sets the article data and updates the viewport content.
func (m *summaryViewModel) SetContent(article *ent.Article, width, height int) tea.Cmd {
	m.article = article
	m.viewport.Width = width
	m.viewport.Height = height - lipgloss.Height(m.headerView()) - lipgloss.Height(m.footerView()) // Adjust height for header/footer
	m.viewport.YPosition = 0                                                                       // Reset scroll position

	content := "Article data not available."
	if article != nil {
		// Combine title, summary, and content for display
		// Use Summary.Summary if available, otherwise Article.Description or Content
		summaryText := article.Description // Default to description
		if article.Edges.Summary != nil && article.Edges.Summary.Summary != "" {
			summaryText = article.Edges.Summary.Summary
		} else if article.Content != "" {
			summaryText = article.Content // Fallback to full content if no summary/description
		}

		content = fmt.Sprintf("\n%s", summaryText)
	}

	m.viewport.SetContent(content)
	m.ready = true // Viewport is ready after content is set
	slog.Debug("Summary view content set", "width", m.viewport.Width, "height", m.viewport.Height, "articleTitle", article.Title)
	return nil
}

func (m summaryViewModel) Init() tea.Cmd {
	slog.Debug("SummaryView model Init called")
	return nil // Content is set via SetContent
}

func (m *summaryViewModel) ShowConfirmationDialog(message string, onYes, onNo func() tea.Cmd) {
	m.showConfirmDialog = true
	m.confirmDialogMsg = message
	m.onConfirmYes = onYes
	m.onConfirmNo = onNo
}

// HideConfirmationDialog hides the confirmation dialog
func (m *summaryViewModel) HideConfirmationDialog() {
	m.showConfirmDialog = false
	m.confirmDialogMsg = ""
	m.onConfirmYes = nil
	m.onConfirmNo = nil
}

func (m summaryViewModel) Update(msg tea.Msg) (summaryViewModel, tea.Cmd) {
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

	slog.Debug("SummaryView model Update called", "msg", msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "b":
			slog.Debug("Back key pressed in summary view")
			return m, func() tea.Msg { return backToArticleListMsg{} } // Send message to main model
		case "j", "down":
			m.viewport.LineDown(1)
		case "k", "up":
			m.viewport.LineUp(1)
		case "g":
			m.viewport.GotoTop()
		case "G":
			m.viewport.GotoBottom()
		case "o":
			if err := OpenArticleURL(m.article.URL); err != nil {
				slog.Error("Failed to open url", "error", err)
			}
		case "r":
			if m.article != nil &&
				m.article.Edges.Feed != nil &&
				!m.article.Edges.Feed.IsBookmark &&
				m.article.Edges.Summary != nil {

				if m.confirm {
					m.ShowConfirmationDialog(
						"記事を既読にしますか？ (y/N)",
						func() tea.Cmd {
							ctx := context.Background()
							return func() tea.Msg {
								if err := m.summaryRepos.UpdateReaded(ctx, m.article.Edges.Summary); err != nil {
									slog.Error("Failed to mark as read", "error", err)
									return errors.Wrap(err, "failed to mark article as read")
								}
								return m
							}
						},
						nil,
					)
					return m, nil
				} else {
					if err := m.summaryRepos.UpdateReaded(context.Background(), m.article.Edges.Summary); err != nil {
						slog.Error("Failed to update summary as readed", "error", err)
					}
					return m, func() tea.Msg { return backToArticleListMsg{} }
				}

			}
		case "p":
			if m.article != nil && m.article.Edges.Summary != nil {
				// Play audio for the summary if available
				slog.Debug("Playing audio for summary")
				go func() {
					sum := m.article.Edges.Summary
					sum.Edges.Feed = m.article.Edges.Feed
					ctx := context.Background()
					audioData, err := summary.GetAudioData(ctx, sum)
					if err != nil {
						slog.Error("Failed to get audio data", "error", err)
						return
					}
					ttsEngine := tts.NewTTSEngine()
					if err := ttsEngine.PlayAudioData(audioData); err != nil {
						slog.Error("Failed to play audio data", "error", err)
					}
				}()
			}
		case "d": // Add delete key binding
			if m.article != nil {
				m.ShowConfirmationDialog(
					"この記事と要約を削除しますか？ (y/N)",
					func() tea.Cmd {
						ctx := context.Background()
						return func() tea.Msg {
							if err := m.articleRepos.Delete(ctx, m.article.ID.String()); err != nil {
								slog.Error("Failed to delete article and summary", "error", err)
								// Optionally return an error message to display to the user
								return errors.Wrap(err, "failed to delete article")
							}
							slog.Debug("Article and summary deleted successfully", "articleID", m.article.ID)
							return backToArticleListMsg{} // Go back to article list after deletion
						}
					},
					nil, // No action on "No"
				)
			}

		}
	case tea.WindowSizeMsg:
		// Only update if ready, otherwise SetContent will handle initial sizing
		if m.ready {
			headerHeight := lipgloss.Height(m.headerView())
			footerHeight := lipgloss.Height(m.footerView())
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - headerHeight - footerHeight
			slog.Debug("Summary view resized", "width", m.viewport.Width, "height", m.viewport.Height)

			go func() {
				if err := m.summaryRepos.UpdateReaded(context.Background(), m.article.Edges.Summary); err != nil {
					slog.Error("Failed to play audio data", "error", err)
				}
			}()

		}
	}

	// Update viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m summaryViewModel) View() string {
	if !m.ready || m.article == nil {
		return "Loading article..."
	}
	content := fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
	if m.showConfirmDialog {
		dialogStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2).
			Width(40)

		dialogBox := dialogStyle.Render(m.confirmDialogMsg)

		return lipgloss.Place(
			m.viewport.Width,
			m.viewport.Height,
			lipgloss.Center,
			lipgloss.Center,
			dialogBox,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
		)

	}
	return content
}

func (m summaryViewModel) headerView() string {
	title := "Article Summary"
	if m.article != nil {
		title = m.article.Title
	}
	if m.article.Edges.Summary != nil && m.article.Edges.Summary.Title != "" {
		title = fmt.Sprintf("%s [%d]\n%s",
			m.article.Edges.Summary.Title,
			len([]rune(m.article.Edges.Summary.Summary)),
			m.article.Title)
	}
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("62")). // Example color
		Padding(0, 1).
		Render(title + "\n\n" + m.article.URL)
}

func (m summaryViewModel) footerView() string {
	// You can add more info here, like scroll percentage
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")). // Dim color
		Padding(0, 1).
		Render("Scroll: ↑/k ↓/j | Top: g | Bottom: G | Play: p | Read: r | Delete: d | Open: o | Back: b ") // Add 'Delete: d'
}
