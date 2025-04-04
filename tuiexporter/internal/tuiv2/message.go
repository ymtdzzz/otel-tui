package tuiv2

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tuiv2/help"
)

const (
	MODAL_TYPE_TRACE = iota
	// MODAL_TYPE_METRIC
	// MODAL_TYPE_LOG
)

type SetTextModalMsg struct {
	Type int
	Text string
}

type ApplySpanFilterMsg struct {
	Query string
}

type UpdateHelpKeysMsg struct{}

func UpdateHelpKeysCmd() tea.Cmd {
	return func() tea.Msg {
		return UpdateHelpKeysMsg{}
	}
}

func SetTraceHelpKeysCmd(
	visible bool,
	id help.ComponentID,
	keys []key.Binding,
) tea.Cmd {
	return func() tea.Msg {
		return help.SetTraceHelpKeysMsg{
			Visible: visible,
			ID:      id,
			Keys:    keys,
		}
	}
}
