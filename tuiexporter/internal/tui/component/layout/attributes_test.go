package layout

import (
	"testing"

	"github.com/rivo/tview"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

func TestAppendAttrsSorted(t *testing.T) {
	tests := []struct {
		name     string
		attrs    map[string]any
		expected []string
	}{
		{
			name:     "empty attributes",
			attrs:    map[string]any{},
			expected: []string{},
		},
		{
			name: "single attribute",
			attrs: map[string]any{
				"service.name": "my-service",
			},
			expected: []string{"service.name: my-service"},
		},
		{
			name: "multiple attributes sorted",
			attrs: map[string]any{
				"zebra":        "animal",
				"apple":        "fruit",
				"service.name": "my-service",
				"banana":       "yellow",
			},
			expected: []string{
				"apple: fruit",
				"banana: yellow",
				"service.name: my-service",
				"zebra: animal",
			},
		},
		{
			name: "different value types",
			attrs: map[string]any{
				"string_val": "hello",
				"int_val":    42,
				"bool_val":   true,
				"float_val":  3.14,
			},
			expected: []string{
				"bool_val: true",
				"float_val: 3.14",
				"int_val: 42",
				"string_val: hello",
			},
		},
		{
			name: "special characters in keys",
			attrs: map[string]any{
				"key with spaces": "value1",
				"key.with.dots":   "value2",
				"key-with-dash":   "value3",
				"key_with_under":  "value4",
			},
			expected: []string{
				"key with spaces: value1",
				"key-with-dash: value3",
				"key.with.dots: value2",
				"key_with_under: value4",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent := tview.NewTreeNode("root")

			attrs := pcommon.NewMap()
			for k, v := range tt.attrs {
				switch val := v.(type) {
				case string:
					attrs.PutStr(k, val)
				case int:
					attrs.PutInt(k, int64(val))
				case bool:
					attrs.PutBool(k, val)
				case float64:
					attrs.PutDouble(k, val)
				}
			}

			AppendAttrsSorted(parent, attrs)

			children := parent.GetChildren()
			if len(children) != len(tt.expected) {
				t.Errorf("expected %d children, got %d", len(tt.expected), len(children))
				return
			}

			for i, child := range children {
				actual := child.GetText()
				if actual != tt.expected[i] {
					t.Errorf("child %d: expected %q, got %q", i, tt.expected[i], actual)
				}
			}
		})
	}
}
