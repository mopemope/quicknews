package tui

import "github.com/charmbracelet/lipgloss"

// Define styles here to be shared across TUI components
var (
	docStyle         = lipgloss.NewStyle().Margin(1, 2)  // Basic margin for the overall view container
	summaryViewStyle = lipgloss.NewStyle().Padding(1, 2) // Style for the summary viewport content area
)

// Example:
// var (
// 	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("62"))
// 	itemStyle = lipgloss.NewStyle().PaddingLeft(2)
// 	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(0).Foreground(lipgloss.Color("62")).SetString("> ")
// 	paginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
// 	helpStyle = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
// )
