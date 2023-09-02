package cli

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

var tableStyle = table.Styles{
	Header: lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("214")).
		BorderBottom(true).
		Bold(true),
}
