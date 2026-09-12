package layout

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/json"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/navigation"
)

func AttachModalForTreeAttributes(tree *tview.TreeView, onHide func()) {
	var currentModalNode *tview.TreeNode = nil
	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		if len(node.GetChildren()) > 0 {
			node.SetExpanded(!node.IsExpanded())
			return
		}
		if currentModalNode == node {
			currentModalNode = nil
			navigation.HideModal(tree)
			if onHide != nil {
				onHide()
			}
			return
		}
		nodeText := node.GetText()
		parts := strings.SplitN(nodeText, ": ", 2)
		if len(parts) >= 2 {
			value := parts[1]
			value = json.PrettyJSON(value)
			nodeText = parts[0] + ": " + value
		}
		textView := navigation.ShowModal(tree, nodeText)
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
			currentModalNode = nil
			navigation.HideModal(tree)
			if onHide != nil {
				onHide()
			}
		}
	})
	tree.SetBlurFunc(func() {
		if currentModalNode == nil {
			return
		}
		currentModalNode = nil
		navigation.QueueModalDismiss(tree, func() {
			if currentModalNode != nil {
				return
			}
			navigation.HideModal(nil)
			if onHide != nil {
				onHide()
			}
		})
	})
}

type tableModalMapper interface {
	// GetColumnIdx returns the column index for getting the content to be shown
	// in the modal
	GetColumnIdx() int
}

func AttachModalForTableRows(table *tview.Table, mapper tableModalMapper, onHide func()) {
	if mapper == nil {
		return
	}

	var currentRow = -1

	table.SetSelectedFunc(func(row, column int) {
		if currentRow == row {
			currentRow = -1
			navigation.HideModal(table)
			if onHide != nil {
				onHide()
			}
			return
		}
		if cell := table.GetCell(row, mapper.GetColumnIdx()); cell != nil {
			text := cell.Text
			text = json.PrettyJSON(text)
			textView := navigation.ShowModal(table, text)
			currentRow = row
			table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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
		}
	})
	table.SetSelectionChangedFunc(func(row, column int) {
		if currentRow != -1 {
			currentRow = -1
			navigation.HideModal(table)
			if onHide != nil {
				onHide()
			}
		}
	})
	table.SetBlurFunc(func() {
		if currentRow == -1 {
			return
		}
		currentRow = -1
		navigation.QueueModalDismiss(table, func() {
			if currentRow != -1 {
				return
			}
			navigation.HideModal(nil)
			if onHide != nil {
				onHide()
			}
		})
	})
}
