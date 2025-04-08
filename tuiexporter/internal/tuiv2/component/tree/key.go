package tree

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tuiv2"
)

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

func (km keyMap) ShortHelp() []key.Binding {
	scrollToTopBottomKeys := km.scrollToTop.Keys()
	scrollToTopBottomKeys = append(scrollToTopBottomKeys, km.scrollToBottom.Keys()...)

	halfPageUpDownKeys := km.halfPageUp.Keys()
	halfPageUpDownKeys = append(halfPageUpDownKeys, km.halfPageDown.Keys()...)

	scrollLeftRightKeys := km.scrollLeft.Keys()
	scrollLeftRightKeys = append(scrollLeftRightKeys, km.scrollRight.Keys()...)

	upDownKeys := km.up.Keys()
	upDownKeys = append(upDownKeys, km.down.Keys()...)

	return []key.Binding{
		key.NewBinding(
			key.WithKeys(scrollToTopBottomKeys...),
			key.WithHelp(
				tuiv2.MergeKeysToString(scrollToTopBottomKeys...),
				"scroll to top/bottom",
			),
		),
		key.NewBinding(
			key.WithKeys(halfPageUpDownKeys...),
			key.WithHelp(
				tuiv2.MergeKeysToString(halfPageUpDownKeys...),
				"scroll up/down ½ page",
			),
		),
		key.NewBinding(
			key.WithKeys(scrollLeftRightKeys...),
			key.WithHelp(
				tuiv2.MergeKeysToString(scrollLeftRightKeys...),
				"scroll left/right",
			),
		),
		key.NewBinding(
			key.WithKeys(upDownKeys...),
			key.WithHelp(
				tuiv2.MergeKeysToString(upDownKeys...),
				"scroll up/down",
			),
		),
		km.selectNode,
	}
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
			key.WithHelp("enter", "expand/collapse, show full text"),
		),
	}
}
