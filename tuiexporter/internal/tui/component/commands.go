package component

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var keyMapRegex = regexp.MustCompile(`Rune|\[|\]`)

type KeyMap struct {
	key         *tcell.EventKey
	arrow       bool
	description string
}

type KeyMaps []*KeyMap

func (m KeyMaps) keyTexts() string {
	keytexts := []string{}
	for _, v := range m {
		key := ""
		if v.arrow {
			key = "→←↑↓"
		} else {
			key = keyMapRegex.ReplaceAllString(v.key.Name(), "")
		}
		keytexts = append(keytexts, fmt.Sprintf("[yellow]%s[white]: %s",
			key,
			v.description,
		))
	}
	return " " + strings.Join(keytexts, " | ")
}

type Focusable interface {
	SetFocusFunc(func()) *tview.Box
}

func newCommandList() *tview.TextView {
	return tview.NewTextView().
		SetDynamicColors(true)
}

func attachCommandList(commands *tview.TextView, p tview.Primitive) *tview.Flex {
	base := tview.NewFlex().SetDirection(tview.FlexRow)

	if commands == nil {
		return base
	}

	base.AddItem(p, 0, 1, true).
		AddItem(commands, 1, 1, false)

	return base
}

func registerCommandList(commands *tview.TextView, c Focusable, origFocusFn func(), keys KeyMaps) {
	if commands == nil {
		return
	}

	c.SetFocusFunc(func() {
		commands.SetText(keys.keyTexts())

		if origFocusFn != nil {
			origFocusFn()
		}
	})
}
