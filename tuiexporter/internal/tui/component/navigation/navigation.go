package navigation

import "github.com/rivo/tview"

var (
	focusFn func(tview.Primitive)
	showMFn func(tview.Primitive, string) *tview.TextView
	hideMFn func(tview.Primitive)
)

func Init(
	setFocusFn func(tview.Primitive),
	showModalFn func(tview.Primitive, string) *tview.TextView,
	hideModalFn func(tview.Primitive),
) {
	focusFn = setFocusFn
	showMFn = showModalFn
	hideMFn = hideModalFn
}

func Focus(primitive tview.Primitive) {
	if focusFn != nil {
		focusFn(primitive)
	}
}

func ShowModal(primitive tview.Primitive, title string) *tview.TextView {
	if showMFn != nil {
		return showMFn(primitive, title)
	}
	return tview.NewTextView()
}

func HideModal(primitive tview.Primitive) {
	if hideMFn != nil {
		hideMFn(primitive)
	}
}
