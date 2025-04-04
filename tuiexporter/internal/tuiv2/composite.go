package tuiv2

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// Position represents a relative offset in the TUI. There are five possible values; Top, Right,
// Bottom, Left, and Center.
type Position int

const (
	Top Position = iota + 1
	Right
	Bottom
	Left
	Center
)

// composite merges and flattens the background and foreground views into a single view.
// This implementation is based on bubbletea-overlay
// https://github.com/rmhubbert/bubbletea-overlay
func Composite(fg, bg string, xPos, yPos Position, xOff, yOff int) string {
	fgWidth, fgHeight := lipgloss.Size(fg)
	bgWidth, bgHeight := lipgloss.Size(bg)

	if fgWidth >= bgWidth && fgHeight >= bgHeight {
		return fg
	}

	x, y := offsets(fg, bg, xPos, yPos, xOff, yOff)
	x = clamp(x, 0, bgWidth-fgWidth)
	y = clamp(y, 0, bgHeight-fgHeight)

	fgLines := lines(fg)
	bgLines := lines(bg)
	var sb strings.Builder

	for i, bgLine := range bgLines {
		if i > 0 {
			sb.WriteByte('\n')
		}
		if i < y || i >= y+fgHeight {
			sb.WriteString(bgLine)
			continue
		}

		pos := 0
		if x > 0 {
			left := ansi.Truncate(bgLine, x, "")
			pos = ansi.StringWidth(left)
			sb.WriteString(left)
			if pos < x {
				sb.WriteString(whitespace(x - pos))
				pos = x
			}
		}

		fgLine := fgLines[i-y]
		sb.WriteString(fgLine)
		pos += ansi.StringWidth(fgLine)

		right := ansi.TruncateLeft(bgLine, pos, "")
		bgWidth := ansi.StringWidth(bgLine)
		rightWidth := ansi.StringWidth(right)
		if rightWidth <= bgWidth-pos {
			sb.WriteString(whitespace(bgWidth - rightWidth - pos))
		}
		sb.WriteString(right)
	}
	return sb.String()
}

// offsets calculates the actual vertical and horizontal offsets used to position the foreground
// tea.Model relative to the background tea.Model.
func offsets(fg, bg string, xPos, yPos Position, xOff, yOff int) (int, int) {
	var x, y int
	switch xPos {
	case Center:
		halfBackgroundWidth := lipgloss.Width(bg) / 2
		halfForegroundWidth := lipgloss.Width(fg) / 2
		x = halfBackgroundWidth - halfForegroundWidth
	case Right:
		x = lipgloss.Width(bg) - lipgloss.Width(fg)
	}

	switch yPos {
	case Center:
		halfBackgroundHeight := lipgloss.Height(bg) / 2
		halfForegroundHeight := lipgloss.Height(fg) / 2
		y = halfBackgroundHeight - halfForegroundHeight
	case Bottom:
		y = lipgloss.Height(bg) - lipgloss.Height(fg)
	}

	// debug(
	// 	"X position: "+strconv.Itoa(int(xPos)),
	// 	"Y position: "+strconv.Itoa(int(yPos)),
	// 	"X offset: "+strconv.Itoa(x+xOff),
	// 	"Y offset: "+strconv.Itoa(y+yOff),
	// 	"Background width: "+strconv.Itoa(lipgloss.Width(bg)),
	// 	"Foreground width: "+strconv.Itoa(lipgloss.Width(fg)),
	// 	"Background height: "+strconv.Itoa(lipgloss.Height(bg)),
	// 	"Foreground height: "+strconv.Itoa(lipgloss.Height(fg)),
	// )

	return x + xOff, y + yOff
}

// clamp calculates the lowest possible number between the given boundaries.
func clamp(v, lower, upper int) int {
	if upper < lower {
		return min(max(v, upper), lower)
	}
	return min(max(v, lower), upper)
}

// lines normalises any non standard new lines within a string, and then splits and returns a slice
// of strings split on the new lines.
func lines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.Split(s, "\n")
}

// whitescpace returns a string of whitespace characters of the requested length.
func whitespace(length int) string {
	return strings.Repeat(" ", length)
}
