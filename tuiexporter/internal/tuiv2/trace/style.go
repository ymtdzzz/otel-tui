package trace

import "github.com/charmbracelet/lipgloss"

type styles struct {
	baseFrame         lipgloss.Style
	focusedFrameColor lipgloss.Color
	left              lipgloss.Style
	right             lipgloss.Style
	modal             lipgloss.Style
}

func defaultStyles() styles {
	baseFrame := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)

	return styles{
		baseFrame:         baseFrame,
		focusedFrameColor: lipgloss.Color("#FFCC00"),
		left:              baseFrame,
		right:             baseFrame,
		modal:             baseFrame,
	}
}

func (s styles) focusedStyle(style lipgloss.Style, focused bool) lipgloss.Style {
	if focused {
		return lipgloss.NewStyle().
			Inherit(style).
			BorderForeground(s.focusedFrameColor)
	}
	return lipgloss.NewStyle().
		Inherit(style).
		BorderForeground(s.baseFrame.GetBorderTopForeground())
}
