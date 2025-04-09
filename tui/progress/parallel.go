package progress

import (
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type parallelProgressModel struct {
	items         []QueueItem
	itemCount     int
	numParallel   int
	index         int
	width         int
	height        int
	spinner       spinner.Model
	progress      progress.Model
	done          bool
	progressLabel string
}

func NewParallelProgressModel(items []QueueItem, progressLabel string, parallel int) parallelProgressModel {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	itemCount := len(items)

	return parallelProgressModel{
		items:         items,
		itemCount:     itemCount,
		numParallel:   parallel,
		spinner:       s,
		progress:      p,
		progressLabel: progressLabel,
	}
}

func (m parallelProgressModel) Init() tea.Cmd {
	return tea.Batch(m.processItem(), m.spinner.Tick)
}

func (m parallelProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		}
	case finishedItemsMsg:

		if m.index >= m.itemCount-1 {
			// Everything's been installed. We're done!
			m.done = true
			return m, tea.Sequence(
				tea.Quit, // exit the program
			)
		}

		m.index += msg.finished

		progressCmd := m.progress.SetPercent(float64(m.index) / float64(m.itemCount))

		return m, tea.Batch(
			progressCmd,
			m.processItem(), // download the next package
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

func (m parallelProgressModel) View() string {
	n := m.itemCount
	w := lipgloss.Width(fmt.Sprintf("%d", n))

	if m.done {
		return doneStyle.Render(fmt.Sprintf("Done %d items.\n", n))
	}

	itemCount := fmt.Sprintf(" %*d/%*d", w, m.index, w, n)

	spin := m.spinner.View() + " "
	prog := m.progress.View()
	cellsAvail := max(0, m.width-lipgloss.Width(spin+prog+itemCount))

	info := lipgloss.NewStyle().MaxWidth(cellsAvail).Render(m.progressLabel + " ")

	cellsRemaining := max(0, m.width-lipgloss.Width(spin+info+prog+itemCount))
	gap := strings.Repeat(" ", cellsRemaining)

	return spin + info + gap + prog + itemCount
}

func (m *parallelProgressModel) processItem() tea.Cmd {
	return func() tea.Msg {
		var wg sync.WaitGroup
		fin := 0

		for range m.numParallel {
			if m.index >= m.itemCount {
				break
			}
			idx := min(m.index+fin, m.itemCount-1)
			item := m.items[idx]
			fin++
			wg.Add(1)
			go func(item QueueItem) {
				defer wg.Done()
				item.Process()
			}(item)
		}
		wg.Wait()

		return finishedItemsMsg{finished: fin}
	}
}

type finishedItemsMsg struct {
	finished int
}
