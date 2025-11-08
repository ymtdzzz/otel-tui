package layout

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var keyMapRegex = regexp.MustCompile(`Rune|\[|\]`)

type KeyMap struct {
	Key         *tcell.EventKey
	Arrow       bool
	Hidden      bool
	Description string
	Handler     func(event *tcell.EventKey) *tcell.EventKey
}

type KeyMaps []*KeyMap

func (m *KeyMaps) Merge(m2 KeyMaps) {
	*m = append(*m, m2...)
}

func (m KeyMaps) keyTexts() string {
	keytexts := []string{}
	for _, v := range m {
		if v.Description == "" || v.Hidden {
			continue
		}
		key := ""
		if v.Arrow {
			key = "→←↑↓"
		} else {
			keyName := v.Key.Name()
			if v.Key.Key() == tcell.KeyCtrlH {
				keyName = "Ctrl-H"
			}
			key = keyMapRegex.ReplaceAllString(keyName, "")
		}
		keytexts = append(keytexts, fmt.Sprintf("[yellow]%s[white]: %s",
			key,
			v.Description,
		))
	}
	return " " + strings.Join(keytexts, " | ")
}

func getInt32Key(key *tcell.EventKey) int32 {
	if key.Key() == tcell.KeyRune {
		return key.Rune() + int32(key.Modifiers())
	}
	return int32(key.Key())
}

type Focusable interface {
	SetFocusFunc(func()) *tview.Box
}

type FocusableBox interface {
	SetFocusFunc(func()) *tview.Box
	SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey) *tview.Box
}

func NewCommandList() *tview.TextView {
	return tview.NewTextView().
		SetDynamicColors(true)
}

func AttachCommandList(commands *tview.TextView, p tview.Primitive) *tview.Flex {
	base := tview.NewFlex().SetDirection(tview.FlexRow)

	if commands == nil {
		return base
	}

	base.AddItem(p, 0, 1, true).
		AddItem(commands, 1, 1, false)

	return base
}

func RegisterCommandList(commands *tview.TextView, c FocusableBox, origFocusFn func(), keys KeyMaps) {
	if commands == nil {
		return
	}

	c.SetFocusFunc(func() {
		commands.SetText(keys.keyTexts())
		log.Printf("triggered SetFocusFunc in RegisterCommandList2. commands: %s\n", commands.GetText(false))

		if origFocusFn != nil {
			origFocusFn()
		}
	})

	km := map[int32]func(event *tcell.EventKey) *tcell.EventKey{}
	for _, k := range keys {
		if k.Handler != nil {
			km[getInt32Key(k.Key)] = k.Handler
		}
	}

	c.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if handler, ok := km[getInt32Key(event)]; ok {
			return handler(event)
		}
		return event
	})
}
