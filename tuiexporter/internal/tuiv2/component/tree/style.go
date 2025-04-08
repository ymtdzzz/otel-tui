package tree

import "github.com/charmbracelet/lipgloss"

type styles struct {
	focusedStyle lipgloss.Style
}

func defaultStyles() styles {
	focusedStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#333333")).
		Bold(true)

	return styles{
		focusedStyle: focusedStyle,
	}
}
