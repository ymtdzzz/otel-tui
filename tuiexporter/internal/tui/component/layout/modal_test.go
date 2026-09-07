package layout

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/navigation"
)

func Test_attachModalForTreeAttributes(t *testing.T) {
	tests := []struct {
		name          string
		nodeText      string
		expectedModal string
	}{
		{
			name:          "simple key-value",
			nodeText:      "key: value",
			expectedModal: "key: value",
		},
		{
			name:     "nested JSON object",
			nodeText: `hoge: {"valid":"value"}`,
			expectedModal: `hoge: {
  "valid": "value"
}`,
		},
		{
			name:     "JSON array",
			nodeText: `array: ["this","is","array"]`,
			expectedModal: `array: [
  "this",
  "is",
  "array"
]`,
		},
		{
			name:          "invalid JSON",
			nodeText:      "hoge: {invalid}",
			expectedModal: "hoge: {invalid}",
		},
		{
			name:          "text with colon but not JSON",
			nodeText:      "time: 12:34:56",
			expectedModal: "time: 12:34:56",
		},
		{
			name:     "nested JSON with multiple colons",
			nodeText: `config: {"time":"12:34:56","url":"http://example.com"}`,
			expectedModal: `config: {
  "time": "12:34:56",
  "url": "http://example.com"
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: Add more assertions to check if the modal is shown correctly
			tree := tview.NewTreeView()
			root := tview.NewTreeNode("")
			node := tview.NewTreeNode(tt.nodeText)
			root.AddChild(node)
			tree.SetRoot(root)

			var modalText string
			showFn := func(_ tview.Primitive, text string) *tview.TextView {
				modalText = text
				return tview.NewTextView()
			}
			hideFn := func(tview.Primitive) {}
			navigation.Init(nil, showFn, hideFn)

			AttachModalForTreeAttributes(tree, nil)
			tree.SetCurrentNode(node)

			event := tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
			handler := tree.InputHandler()
			handler(event, nil)

			assert.Equal(t, tt.expectedModal, modalText)
		})
	}
}

func TestAttachModalForTreeAttributesHidesModalOnBlur(t *testing.T) {
	tree := tview.NewTreeView()
	root := tview.NewTreeNode("")
	node := tview.NewTreeNode("key: value")
	root.AddChild(node)
	tree.SetRoot(root).SetCurrentNode(node)

	hideCalls := 0
	onHideCalls := 0
	navigation.Init(nil, func(tview.Primitive, string) *tview.TextView {
		return tview.NewTextView()
	}, func(current tview.Primitive) {
		assert.Nil(t, current)
		hideCalls++
	})

	AttachModalForTreeAttributes(tree, func() { onHideCalls++ })
	tree.GetSelectedFunc()(node)
	tree.Blur()

	assert.Equal(t, 1, hideCalls)
	assert.Equal(t, 1, onHideCalls)
}

type testTableModalMapper struct{}

func (testTableModalMapper) GetColumnIdx() int { return 0 }

func TestAttachModalForTableRowsHidesModalOnBlur(t *testing.T) {
	table := tview.NewTable().SetSelectable(true, false)
	table.SetCell(0, 0, tview.NewTableCell("value"))

	hideCalls := 0
	onHideCalls := 0
	navigation.Init(nil, func(tview.Primitive, string) *tview.TextView {
		return tview.NewTextView()
	}, func(current tview.Primitive) {
		assert.Nil(t, current)
		hideCalls++
	})

	AttachModalForTableRows(table, testTableModalMapper{}, func() { onHideCalls++ })
	// Invoke the selected handler through the input path without allowing the
	// same event's selection notification to close the modal first.
	table.SetSelectionChangedFunc(nil)
	table.Focus(nil)
	table.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), nil)
	table.Blur()

	assert.Equal(t, 1, hideCalls)
	assert.Equal(t, 1, onHideCalls)
}
