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

// keyID identifies a key binding. tcell.Key and rune values share the same
// numeric space (e.g. tcell.KeyCtrlL and 'L' are both 76), so they must be kept
// in separate fields to avoid collisions.
type keyID struct {
	key tcell.Key
	ch  rune
	mod tcell.ModMask
}

func getKeyID(key *tcell.EventKey) keyID {
	// Rune keys are identified by ch and mod, since Key() is always tcell.KeyRune.
	if key.Key() == tcell.KeyRune {
		return keyID{key: tcell.KeyRune, ch: key.Rune(), mod: key.Modifiers()}
	}
	// Other keys are identified by Key() alone, because ch and mod don't match
	// between a registered binding and a real event. NewEventKey(KeyEnter, ' ',
	// ModNone) keeps ch=' ' while a terminal sends ch=0, and Ctrl keys arrive
	// with ModCtrl although bindings are written with ModNone.
	return keyID{key: key.Key()}
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
		log.Printf("triggered SetFocusFunc in RegisterCommandList. commands: %s\n", commands.GetText(false))

		if origFocusFn != nil {
			origFocusFn()
		}
	})

	km := map[keyID]func(event *tcell.EventKey) *tcell.EventKey{}
	for _, k := range keys {
		if k.Handler != nil {
			km[getKeyID(k.Key)] = k.Handler
		}
	}

	c.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if handler, ok := km[getKeyID(event)]; ok {
			return handler(event)
		}
		return event
	})
}
