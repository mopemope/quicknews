package progress

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"
)

type countingItem struct {
	name  string
	url   string
	calls atomic.Int64
	err   error
	delay time.Duration
}

func (i *countingItem) DisplayName() string { return i.name }
func (i *countingItem) URL() string         { return i.url }
func (i *countingItem) Process() error {
	i.calls.Add(1)
	if i.delay > 0 {
		time.Sleep(i.delay)
	}
	return i.err
}

// collect executes a command chain the way the bubbletea runtime would,
// expanding batch commands into their resulting messages.
func collect(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	msg := cmd()
	switch m := msg.(type) {
	case nil:
		return nil
	case tea.BatchMsg:
		var out []tea.Msg
		for _, c := range m {
			out = append(out, collect(c)...)
		}
		return out
	default:
		return []tea.Msg{m}
	}
}

// runParallelModel drives the model's update loop until all work is done.
func runParallelModel(t *testing.T, m parallelProgressModel) parallelProgressModel {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	msgs := collect(m.Init())
	for len(msgs) > 0 {
		if ctx.Err() != nil {
			t.Fatal("timed out waiting for progress model to finish")
		}
		msg := msgs[0]
		msgs = msgs[1:]

		switch msg.(type) {
		case tea.QuitMsg:
			return m
		case spinner.TickMsg:
			// Drop spinner ticks; they would otherwise loop forever.
			continue
		}

		var next tea.Cmd
		var model tea.Model
		model, next = m.Update(msg)
		m = model.(parallelProgressModel)
		msgs = append(msgs, collect(next)...)
	}
	return m
}

func TestParallelProgressModel_ProcessesEachItemExactlyOnce(t *testing.T) {
	const total = 7
	items := make([]QueueItem, 0, total)
	for i := range total {
		items = append(items, &countingItem{
			name: fmt.Sprintf("item-%d", i),
			url:  fmt.Sprintf("https://example.com/%d", i),
			// Small jitter so goroutines interleave like a real run.
			delay: time.Millisecond * time.Duration(1+i%3),
		})
	}

	m := NewParallelProgressModel(items, "Testing", 5)
	m = runParallelModel(t, m)

	require.True(t, m.done)
	for i, item := range items {
		ci := item.(*countingItem)
		require.EqualValues(t, 1, ci.calls.Load(), "item %d processed %d times", i, ci.calls.Load())
	}
	require.EqualValues(t, total, m.index)
}

func TestParallelProgressModel_HandlesErrorItems(t *testing.T) {
	const total = 4
	items := make([]QueueItem, 0, total)
	for i := range total {
		item := &countingItem{
			name: fmt.Sprintf("item-%d", i),
			url:  fmt.Sprintf("https://example.com/%d", i),
		}
		if i == 1 {
			item.err = fmt.Errorf("boom-%d", i)
		}
		items = append(items, item)
	}

	m := NewParallelProgressModel(items, "Testing", 3)
	m = runParallelModel(t, m)

	require.True(t, m.done)
	require.Error(t, m.Err())
	require.ErrorContains(t, m.Err(), "boom-1")
	for i, item := range items {
		ci := item.(*countingItem)
		require.EqualValues(t, 1, ci.calls.Load(), "item %d processed %d times", i, ci.calls.Load())
	}
}

func TestParallelProgressModel_HandlesEmptyItems(t *testing.T) {
	m := NewParallelProgressModel(nil, "Testing", 5)
	m = runParallelModel(t, m)
	require.True(t, m.done)
	require.NoError(t, m.Err())
}
