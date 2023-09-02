package cli

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

var (
	tableStyle = table.Styles{
		Header: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("214")).
			BorderBottom(true).
			Bold(true),
	}

	keyStyle = lipgloss.NewStyle().
			Width(32).
			AlignHorizontal(lipgloss.Right).
			MarginRight(2)

	valueStyle = lipgloss.NewStyle()

	sectionHeaderStyle = lipgloss.NewStyle()
)

func kvPrintln(key string, value string) {
	print(keyStyle.Render(key))
	println(valueStyle.Render(value))
}
