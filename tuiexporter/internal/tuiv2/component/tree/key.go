package tree

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	scrollToTop    key.Binding
	scrollToBottom key.Binding
	halfPageUp     key.Binding
	halfPageDown   key.Binding
	scrollRight    key.Binding
	scrollLeft     key.Binding
	down           key.Binding
	up             key.Binding
	selectNode     key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		scrollToTop: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "scroll to top"),
		),
		scrollToBottom: key.NewBinding(
			key.WithKeys("G"),
			key.WithHelp("G", "scroll to bottom"),
		),
		halfPageUp: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("^u", "½ page up"),
		),
		halfPageDown: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("^d", "½ page down"),
		),
		scrollRight: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("right/l", "scroll right"),
		),
		scrollLeft: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("left/h", "scroll left"),
		),
		down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("down/j", "down"),
		),
		up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("up/k", "up"),
		),
		selectNode: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "expand/collapse node"),
		),
	}
}
