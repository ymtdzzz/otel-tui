package layout

import (
	"fmt"
	"sort"

	"github.com/rivo/tview"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

// AppendAttrsSorted appends attributes to the given parent node in sorted order by key.
func AppendAttrsSorted(parent *tview.TreeNode, attrs pcommon.Map) {
	keys := make([]string, 0, attrs.Len())
	attrs.Range(func(k string, _ pcommon.Value) bool {
		keys = append(keys, k)
		return true
	})
	sort.Strings(keys)

	for _, k := range keys {
		v, _ := attrs.Get(k)
		attr := tview.NewTreeNode(fmt.Sprintf("%s: %s", k, v.AsString()))
		parent.AddChild(attr)
	}
}
