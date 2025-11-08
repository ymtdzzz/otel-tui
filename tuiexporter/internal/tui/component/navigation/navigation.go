package navigation

import "github.com/rivo/tview"

var (
	focusFn func(tview.Primitive)
)

func Init(
	setFocusFn func(tview.Primitive),
) {
	focusFn = setFocusFn
}

func Focus(primitive tview.Primitive) {
	if focusFn != nil {
		focusFn(primitive)
	}
}
