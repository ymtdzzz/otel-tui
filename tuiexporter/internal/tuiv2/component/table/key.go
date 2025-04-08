package table

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	focusInput key.Binding
	applyInput key.Binding
	blurInput  key.Binding
}

func (km keyMap) ShortHelp(isInputFocused bool) []key.Binding {
	if isInputFocused {
		return []key.Binding{
			km.applyInput,
			km.blurInput,
		}
	}
	return []key.Binding{
		km.focusInput,
	}
}

func defaultKeyMap() keyMap {
	return keyMap{
		focusInput: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "focus query input"),
		),
		applyInput: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "apply query input"),
		),
		blurInput: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", " cancel query input"),
		),
	}
}
