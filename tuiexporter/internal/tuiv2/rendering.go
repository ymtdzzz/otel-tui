package tuiv2

import (
	"strings"

	"github.com/acarl005/stripansi"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

func RenderWithTitle(view, title string, style lipgloss.Style) string {
	lines := strings.Split(view, "\n")
	if len(lines) == 0 {
		return view
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(style.GetBorderTopForeground())

	plain := stripansi.Strip(lines[0])
	var b strings.Builder
	titleWidth := runewidth.StringWidth(title) + 2
	for i := 0; i < titleWidth; i++ {
		b.WriteString(style.GetBorderStyle().Top)
	}
	replaced := strings.Replace(plain, b.String(), " "+title+" ", 1)

	lines[0] = titleStyle.Render(replaced)

	return strings.Join(lines, "\n")
}
