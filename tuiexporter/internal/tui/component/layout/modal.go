package layout

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/json"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/navigation"
)

func AttachModalForTreeAttributes(tree *tview.TreeView) {
	var currentModalNode *tview.TreeNode = nil
	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		if len(node.GetChildren()) > 0 {
			node.SetExpanded(!node.IsExpanded())
			return
		}
		if currentModalNode == node {
			navigation.HideModal(tree)
			currentModalNode = nil
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
			navigation.HideModal(tree)
			currentModalNode = nil
		}
	})
}

type tableModalMapper interface {
	// GetColumnIdx returns the column index for getting the content to be shown
	// in the modal
	GetColumnIdx() int
}

func AttachModalForTableRows(table *tview.Table, mapper tableModalMapper) {
	if mapper == nil {
		return
	}

	var currentRow = -1

	table.SetSelectedFunc(func(row, column int) {
		if currentRow == row {
			navigation.HideModal(table)
			currentRow = -1
			return
		}
		currentRow = row
		if cell := table.GetCell(row, mapper.GetColumnIdx()); cell != nil {
			text := cell.Text
			text = json.PrettyJSON(text)
			textView := navigation.ShowModal(table, text)
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
			navigation.HideModal(table)
			currentRow = -1
		}
	})
}
