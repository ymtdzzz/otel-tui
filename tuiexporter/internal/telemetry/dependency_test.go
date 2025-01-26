package telemetry

import (
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

func TestGetMermaid(t *testing.T) {
	t.Run("Head only nodes", func(t *testing.T) {
		sdm := SpanDataMap{}
		addSpan(t, sdm, 1, 1, "serviceA", "")
		addSpan(t, sdm, 2, 2, "serviceB", "")
		addSpan(t, sdm, 3, 3, "serviceC", "")

		gotstr := sdm.getDependencies().getMermaid()
		wantstr := `graph LR
serviceA
serviceB
serviceC
`
		got := strings.Split(gotstr, "\n")
		want := strings.Split(wantstr, "\n")

		assert.Equal(t, want[0], got[0])
		assert.ElementsMatch(t, want[1:], got[1:])
	})
	t.Run("Single relation", func(t *testing.T) {
		sdm := SpanDataMap{}
		addSpan(t, sdm, 1, 1, "serviceA", "serviceB")
		addSpan(t, sdm, 2, 3, "serviceA", "serviceB")
		addSpan(t, sdm, 3, 5, "serviceA", "serviceB")

		got := sdm.getDependencies().getMermaid()
		want := `graph LR
serviceA -->|3| serviceB
`
		assert.Equal(t, want, got)
	})
	t.Run("Complicated", func(t *testing.T) {
		sdm := SpanDataMap{}
		// Head only node
		addSpan(t, sdm, 1, 1, "serviceS", "")
		// Depth
		addSpan(t, sdm, 2, 2, "serviceA", "serviceB")
		addSpan(t, sdm, 3, 4, "serviceB", "serviceC")
		addSpan(t, sdm, 4, 6, "serviceC", "serviceD")
		// Multiple children
		addSpan(t, sdm, 3, 8, "serviceB", "serviceE")
		addSpan(t, sdm, 5, 10, "serviceB", "serviceF")
		// Count up
		addSpan(t, sdm, 6, 12, "serviceB", "serviceC")
		addSpan(t, sdm, 7, 14, "serviceB", "serviceC")
		// Other
		addSpan(t, sdm, 8, 16, "serviceX", "serviceY")

		deps := sdm.getDependencies()
		// sort to avoid flaky
		for _, n := range deps.HeadNodes {
			if n.Service == "serviceA" {
				bn := n.Children[0]
				bchild := bn.Children
				sort.Slice(bchild, func(i, j int) bool {
					return bchild[i].Service < bchild[j].Service
				})
			}
		}
		gotstr := deps.getMermaid()
		wantstr := `graph LR
serviceA -->|1| serviceB -->|3| serviceC -->|1| serviceD
serviceB -->|1| serviceE
serviceB -->|1| serviceF
serviceX -->|1| serviceY
serviceS
`
		got := strings.Split(gotstr, "\n")
		want := strings.Split(wantstr, "\n")

		assert.Equal(t, want[0], got[0])
		assert.ElementsMatch(t, want[1:], got[1:])
	})
}

func TestGetSortedMermaid(t *testing.T) {
	input := `graph LR
serviceH -->|1| serviceI -->|3| serviceJ
serviceX -->|1| serviceY
serviceS
serviceA -->|1| serviceB -->|3| serviceC -->|1| serviceD
`
	want := `graph LR
serviceA -->|1| serviceB -->|3| serviceC -->|1| serviceD
serviceH -->|1| serviceI -->|3| serviceJ
serviceX -->|1| serviceY
serviceS
`

	got := getSortedMermaid(input)

	assert.Equal(t, want, got)
}

func addSpan(t *testing.T, sdm SpanDataMap, traceID, spanID int, fromsn, tosn string) {
	t.Helper()

	frs := ptrace.NewResourceSpans()
	frs.Resource().Attributes().PutStr("service.name", fromsn)
	frs.ScopeSpans().EnsureCapacity(1)
	fss := frs.ScopeSpans().AppendEmpty()
	fss.Spans().EnsureCapacity(1)
	fs := fss.Spans().AppendEmpty()
	fs.SetTraceID([16]byte{byte(traceID)})
	fs.SetSpanID([8]byte{byte(spanID)})

	from := &SpanData{
		Span:         &fs,
		ResourceSpan: &frs,
		ScopeSpans:   &fss,
		ReceivedAt:   time.Now(),
	}
	sdm[fs.SpanID().String()] = from

	if len(tosn) == 0 {
		return
	}

	trs := ptrace.NewResourceSpans()
	trs.Resource().Attributes().PutStr("service.name", tosn)
	trs.ScopeSpans().EnsureCapacity(1)
	tss := trs.ScopeSpans().AppendEmpty()
	tss.Spans().EnsureCapacity(1)
	ts := tss.Spans().AppendEmpty()
	ts.SetTraceID([16]byte{byte(traceID)})
	ts.SetSpanID([8]byte{byte(spanID + 1)})
	ts.SetParentSpanID(fs.SpanID())

	to := &SpanData{
		Span:         &ts,
		ResourceSpan: &trs,
		ScopeSpans:   &tss,
		ReceivedAt:   time.Now(),
	}
	sdm[ts.SpanID().String()] = to
}
