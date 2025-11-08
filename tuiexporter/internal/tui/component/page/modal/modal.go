package modal

import (
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/navigation"
)

const ModalTitle = "Scroll (Ctrl+J, Ctrl+K)"

type ModalPage struct {
	view     *tview.Flex
	textView *tview.TextView
}

func NewModalPage() *ModalPage {
	textView := tview.NewTextView()
	textView.SetBorder(true).SetTitle(ModalTitle)

	container := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 2, false).
		AddItem(nil, 0, 2, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 2, false).
			AddItem(nil, 0, 1, false).
			AddItem(textView, 0, 1, false), 0, 3, false)

	return &ModalPage{
		view:     container,
		textView: textView,
	}
}

func (m *ModalPage) SetText(text string) {
	m.textView.SetText(text)
}

func (m *ModalPage) GetPrimitive() tview.Primitive {
	return m.view
}

func (m *ModalPage) ShowModalFunc(showModalPageFn func()) layout.ShowModalFunc {
	return func(current tview.Primitive, text string) *tview.TextView {
		m.SetText(text)
		showModalPageFn()
		navigation.Focus(current)
		return m.textView
	}
}

func (m *ModalPage) HideModalFunc(hideModalPageFn func()) layout.HideModalFunc {
	return func(current tview.Primitive) {
		hideModalPageFn()
		navigation.Focus(current)
	}
}
