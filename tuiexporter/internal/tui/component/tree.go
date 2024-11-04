package component

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type showModalFunc func(tview.Primitive, string) *tview.TextView

type hideModalFunc func(tview.Primitive)

func attachModalForTreeAttributes(tree *tview.TreeView, showFn showModalFunc, hideFn hideModalFunc) {
	var currentModalNode *tview.TreeNode = nil
	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		if len(node.GetChildren()) > 0 {
			node.SetExpanded(!node.IsExpanded())
			return
		}
		if currentModalNode == node {
			hideFn(tree)
			currentModalNode = nil
			return
		}
		textView := showFn(tree, node.GetText())
		textView.SetTitle("Scroll (Ctrl+J, Ctrl+K)")
		currentModalNode = node
		tree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyCtrlJ:
				row, col := textView.GetScrollOffset()
				textView.ScrollTo(row+1, col)
				return nil
			case tcell.KeyCtrlK:
				row, col := textView.GetScrollOffset()
				textView.ScrollTo(row-1, col)
				return nil
			}
			return event
		})
	})
	tree.SetChangedFunc(func(node *tview.TreeNode) {
		if currentModalNode != nil {
			hideFn(tree)
			currentModalNode = nil
		}
	})
}
