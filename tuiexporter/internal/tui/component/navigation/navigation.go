package navigation

import "github.com/rivo/tview"

var (
	focusFn         func(tview.Primitive)
	showMFn         func(tview.Primitive, string) *tview.TextView
	hideMFn         func(tview.Primitive)
	queueUpdateFn   func(func())
	modalOwner      tview.Primitive
	modalGeneration uint64
)

func Init(
	setFocusFn func(tview.Primitive),
	showModalFn func(tview.Primitive, string) *tview.TextView,
	hideModalFn func(tview.Primitive),
) {
	focusFn = setFocusFn
	showMFn = showModalFn
	hideMFn = hideModalFn
	queueUpdateFn = nil
	modalOwner = nil
	modalGeneration = 0
}

// SetQueueUpdateFunc schedules work after tview releases its focus lock.
func SetQueueUpdateFunc(queue func(func())) {
	queueUpdateFn = queue
}

// QueueModalDismiss ignores a delayed blur if a newer modal has replaced it.
func QueueModalDismiss(owner tview.Primitive, update func()) {
	if modalOwner != owner {
		return
	}
	generation := modalGeneration
	guarded := func() {
		if generation == modalGeneration && modalOwner == owner {
			update()
		}
	}
	if queueUpdateFn != nil {
		queueUpdateFn(guarded)
		return
	}
	guarded()
}

func Focus(primitive tview.Primitive) {
	if focusFn != nil {
		focusFn(primitive)
	}
}

func ShowModal(primitive tview.Primitive, title string) *tview.TextView {
	modalOwner = primitive
	modalGeneration++
	if showMFn != nil {
		return showMFn(primitive, title)
	}
	return tview.NewTextView()
}

func HideModal(primitive tview.Primitive) {
	modalOwner = nil
	modalGeneration++
	if hideMFn != nil {
		hideMFn(primitive)
	}
}
