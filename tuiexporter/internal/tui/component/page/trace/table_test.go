package trace

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/test"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

func traceAt(t *testing.T, traceID int, start time.Time) ptrace.Traces {
	t.Helper()
	payload, _ := test.GenerateOTLPTracesPayload(t, traceID, 1, []int{1}, [][]int{{1}})
	span := payload.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0)
	span.SetStartTimestamp(pcommon.NewTimestampFromTime(start))
	span.SetEndTimestamp(pcommon.NewTimestampFromTime(start.Add(time.Millisecond)))
	return payload
}

func TestTableKeepsSelectedTraceWhenOlderTracesArrive(t *testing.T) {
	_, page, screen, store := setupTracePage(t)
	base := time.Date(2025, 11, 9, 12, 0, 0, 0, time.UTC)

	newer := traceAt(t, 1, base.Add(2*time.Second))
	store.AddSpan(&newer)
	test.Render(t, page.view, screen)
	row, _ := page.table.table.GetSelection()
	assert.Equal(t, 1, row)

	older := traceAt(t, 2, base.Add(time.Second))
	store.AddSpan(&older)
	test.Render(t, page.view, screen)

	row, _ = page.table.table.GetSelection()
	assert.Equal(t, 2, row)
	assert.Equal(t, newer.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).TraceID().String(), store.GetTraceIDByFilteredIdx(row-1))
}
