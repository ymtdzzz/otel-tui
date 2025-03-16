package component

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func prettyJSON(s string) string {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(s), "", "  "); err != nil {
		return s
	}
	return prettyJSON.String()
}

type showModalFunc func(tview.Primitive, string) *tview.TextView

type hideModalFunc func(tview.Primitive)

const MODAL_TITLE = "Scroll (Ctrl+J, Ctrl+K)"

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
		nodeText := node.GetText()
		parts := strings.SplitN(nodeText, ": ", 2)
		if len(parts) >= 2 {
			value := parts[1]
			if json.Valid([]byte(value)) {
				value = prettyJSON(value)
			}
			nodeText = parts[0] + ": " + value
		}
		textView := showFn(tree, nodeText)
		textView.SetTitle(MODAL_TITLE)
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

type tableModalMapper interface {
	// GetColumnIdx returns the column index for getting the content to be shown
	// in the modal
	GetColumnIdx() int
}

func attachModalForTableRows(table *tview.Table, mapper tableModalMapper, showFn showModalFunc, hideFn hideModalFunc) {
	if mapper == nil {
		return
	}

	var currentRow = -1

	table.SetSelectedFunc(func(row, column int) {
		if currentRow == row {
			hideFn(table)
			currentRow = -1
			return
		}
		currentRow = row
		if cell := table.GetCell(row, mapper.GetColumnIdx()); cell != nil {
			text := cell.Text
			if json.Valid([]byte(text)) {
				text = prettyJSON(text)
			}
			textView := showFn(table, text)
			textView.SetTitle(MODAL_TITLE)
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
			hideFn(table)
			currentRow = -1
		}
	})
}
