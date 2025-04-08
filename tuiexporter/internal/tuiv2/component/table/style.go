package table

import "github.com/charmbracelet/lipgloss"

type styles struct {
	headerBorderColor lipgloss.Color
}

func defaultStyles() styles {
	headerBottomBorderColor := lipgloss.Color("240")

	return styles{
		headerBorderColor: headerBottomBorderColor,
	}
}
