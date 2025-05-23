package progress

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mopemope/quicknews/config"
	"github.com/mopemope/quicknews/ent"
	"github.com/mopemope/quicknews/models/bookmark"
	"github.com/mopemope/quicknews/tui"
)

type QueueItem interface {
	DisplayName() string
	URL() string
	Process()
}

type Config struct {
	Client        *ent.Client
	Config        *config.Config
	Items         []QueueItem
	ProgressLabel string
}

type singleProgressModel struct {
	items         []QueueItem
	index         int
	width         int
	height        int
	spinner       spinner.Model
	progress      progress.Model
	done          bool
	progressLabel string
	bookmarkRepos bookmark.Repository
}

var (
	currentPkgNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	doneStyle           = lipgloss.NewStyle().Margin(1, 2)
	checkMark           = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
)

func NewSingleProgressModel(ctx context.Context, config *Config) singleProgressModel {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	bookmarkRepos, _ := bookmark.NewRepository(ctx, config.Client, config.Config)

	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	return singleProgressModel{
		items:         config.Items,
		spinner:       s,
		progress:      p,
		progressLabel: config.ProgressLabel,
		bookmarkRepos: bookmarkRepos,
	}
}

func (m singleProgressModel) Init() tea.Cmd {
	return tea.Batch(m.processItem(m.items[m.index]), m.spinner.Tick)
}

func (m singleProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		case "B":
			if m.bookmarkRepos != nil {
				if err := m.bookmarkRepos.AddBookmark(context.Background(), m.items[m.index].URL()); err != nil {
					slog.Error("Failed to add bookmark", "error", err)
				} else {
					slog.Info("Bookmark added", "url", m.items[m.index].URL())
				}
			} else {
				slog.Warn("Bookmark repository not initialized")
			}
		case "o":
			if err := tui.OpenArticleURL(m.items[m.index].URL()); err != nil {
				slog.Error("Failed to open url", "error", err)
			}
		}
	case finishedItemMsg:
		name := m.items[m.index].DisplayName()
		if m.index >= len(m.items)-1 {
			// Everything's been installed. We're done!
			m.done = true
			return m, tea.Sequence(
				tea.Printf("%s %s", checkMark, name), // print the last success message
				tea.Quit,                             // exit the program
			)
		}

		// Update progress bar
		m.index++
		progressCmd := m.progress.SetPercent(float64(m.index) / float64(len(m.items)))

		return m, tea.Batch(
			progressCmd,
			tea.Printf("%s %s", checkMark, name), // print success message above our program
			m.processItem(m.items[m.index]),      // download the next package
		)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case progress.FrameMsg:
		newModel, cmd := m.progress.Update(msg)
		if newModel, ok := newModel.(progress.Model); ok {
			m.progress = newModel
		}
		return m, cmd
	}
	return m, nil
}

func (m singleProgressModel) View() string {
	n := len(m.items)
	w := lipgloss.Width(fmt.Sprintf("%d", n))

	if m.done {
		return doneStyle.Render(fmt.Sprintf("Done %d items.\n", n))
	}

	itemCount := fmt.Sprintf(" %*d/%*d", w, m.index, w, n)

	spin := m.spinner.View() + " "
	prog := m.progress.View()
	cellsAvail := max(0, m.width-lipgloss.Width(spin+prog+itemCount))

	itemName := currentPkgNameStyle.Render(m.items[m.index].DisplayName())
	info := lipgloss.NewStyle().MaxWidth(cellsAvail).Render(m.progressLabel + " " + itemName)

	cellsRemaining := max(0, m.width-lipgloss.Width(spin+info+prog+itemCount))
	gap := strings.Repeat(" ", cellsRemaining)

	return spin + info + gap + prog + itemCount
}

func (m *singleProgressModel) processItem(item QueueItem) tea.Cmd {
	return func() tea.Msg {
		item.Process()
		return finishedItemMsg{item: item}
	}
}

type finishedItemMsg struct {
	item QueueItem
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
