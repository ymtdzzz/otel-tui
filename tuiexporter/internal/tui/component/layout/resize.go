package layout

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ResizeDirection int

const (
	ResizeDirectionHorizontal ResizeDirection = iota
	ResizeDirectionVertical

	WidenHorizontallyKey  = tcell.KeyCtrlL
	NarrowHorizontallyKey = tcell.KeyCtrlH
	WidenVerticallyKey    = tcell.KeyCtrlJ
	NarrowVerticallyKey   = tcell.KeyCtrlK
)

type ResizeManager struct {
	direction        ResizeDirection
	parent           *tview.Flex
	first, second    tview.Primitive
	firstProportion  int
	secondProportion int
	commands         *tview.TextView
}

func NewResizeManager(direction ResizeDirection) *ResizeManager {
	return &ResizeManager{
		direction: direction,
	}
}

func (m *ResizeManager) Register(
	parent *tview.Flex,
	first, second tview.Primitive,
	firstProportion, secondProportion int,
	commands *tview.TextView,
) {
	m.firstProportion = firstProportion
	m.secondProportion = secondProportion
	m.parent = parent
	m.first = first
	m.second = second
	m.commands = commands
}

func (m *ResizeManager) KeyMaps() KeyMaps {
	switch m.direction {
	case ResizeDirectionHorizontal:
		moveDividerLeft := func(event *tcell.EventKey) *tcell.EventKey {
			if m.firstProportion <= 1 {
				return nil
			}
			m.firstProportion--
			m.secondProportion++
			m.resize()
			return nil
		}
		return KeyMaps{
			{
				Key:         tcell.NewEventKey(tcell.KeyCtrlH, ' ', tcell.ModNone),
				Description: "Move divider left",
				Handler:     moveDividerLeft,
			},
			{
				// Ctrl-H is often interpreted as backspace by terminals
				Key:     tcell.NewEventKey(tcell.KeyBackspace, ' ', tcell.ModNone),
				Hidden:  true,
				Handler: moveDividerLeft,
			},
			{
				// Some terminals use KeyBackspace2 for backspace
				Key:     tcell.NewEventKey(tcell.KeyBackspace2, ' ', tcell.ModNone),
				Hidden:  true,
				Handler: moveDividerLeft,
			},
			{
				Key:         tcell.NewEventKey(tcell.KeyCtrlL, ' ', tcell.ModNone),
				Description: "Move divider right",
				Handler: func(event *tcell.EventKey) *tcell.EventKey {
					if m.secondProportion <= 1 {
						return nil
					}
					m.firstProportion++
					m.secondProportion--
					m.resize()
					return nil
				},
			},
		}
	case ResizeDirectionVertical:
		return KeyMaps{
			{
				Key:         tcell.NewEventKey(tcell.KeyCtrlJ, ' ', tcell.ModNone),
				Description: "Move divider down",
				Handler: func(event *tcell.EventKey) *tcell.EventKey {
					if m.secondProportion <= 1 {
						return nil
					}
					m.firstProportion++
					m.secondProportion--
					m.resize()
					return nil
				},
			},
			{
				Key:         tcell.NewEventKey(tcell.KeyCtrlK, ' ', tcell.ModNone),
				Description: "Mode divider up",
				Handler: func(event *tcell.EventKey) *tcell.EventKey {
					if m.firstProportion <= 1 {
						return nil
					}
					m.firstProportion--
					m.secondProportion++
					m.resize()
					return nil
				},
			},
		}
	}
	return nil
}

func (m *ResizeManager) resize() {
	m.parent.ResizeItem(m.first, 0, m.firstProportion).
		ResizeItem(m.second, 0, m.secondProportion)
}
