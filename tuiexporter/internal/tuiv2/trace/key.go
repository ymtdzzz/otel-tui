package trace

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tuiv2"
)

type keyMap struct {
	focusTraces      key.Binding
	focusDetails     key.Binding
	moveSplitRight   key.Binding
	moveSplitLeft    key.Binding
	modalScrollUp    key.Binding
	modalScrollDown  key.Binding
	modalScrollRight key.Binding
	modalScrollLeft  key.Binding
}

func (km keyMap) ShortHelp() []key.Binding {
	moveSplitKeys := []string{
		km.moveSplitLeft.Keys()[0],
		km.moveSplitRight.Keys()[0],
	}

	return []key.Binding{
		key.NewBinding(
			key.WithKeys(moveSplitKeys...),
			key.WithHelp(
				tuiv2.MergeKeysToString(moveSplitKeys...),
				"widen or narrow width",
			),
		),
	}
}

func defaultKeyMap() keyMap {
	return keyMap{
		focusTraces: key.NewBinding(
			key.WithKeys("t"),
		),
		focusDetails: key.NewBinding(
			key.WithKeys("d"),
		),
		moveSplitRight: key.NewBinding(
			key.WithKeys("ctrl+l"),
			key.WithHelp("^l", "widen or narrow width"),
		),
		moveSplitLeft: key.NewBinding(
			key.WithKeys("ctrl+h"),
			key.WithHelp("^h", "widen or narrow width"),
		),
		modalScrollUp: key.NewBinding(
			key.WithKeys("ctrl+k"),
		),
		modalScrollDown: key.NewBinding(
			key.WithKeys("ctrl+j"),
		),
		modalScrollRight: key.NewBinding(
			key.WithKeys("ctrl+l"),
		),
		modalScrollLeft: key.NewBinding(
			key.WithKeys("ctrl+h"),
		),
	}
}
