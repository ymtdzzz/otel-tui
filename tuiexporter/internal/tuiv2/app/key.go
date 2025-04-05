package app

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	quit         key.Binding
	changeTabKey key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("^c", "Quit"),
		),
		changeTabKey: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "Change tab"),
		),
	}
}
