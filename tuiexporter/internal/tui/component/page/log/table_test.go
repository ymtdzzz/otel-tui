package log

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/test"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
)

func logAt(t *testing.T, traceID int, ts time.Time) plog.Logs {
	t.Helper()
	payload, _ := test.GenerateOTLPLogsPayload(t, traceID, 1, []int{1}, [][]int{{1}})
	lrs := payload.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords()
	for i := 0; i < lrs.Len(); i++ {
		lrs.At(i).SetTimestamp(pcommon.NewTimestampFromTime(ts))
	}
	return payload
}

func TestTableKeepsSelectedLogWhenOlderLogsArrive(t *testing.T) {
	_, page, screen, store := setupLogPage(t)
	base := time.Date(2025, 11, 9, 12, 0, 0, 0, time.UTC)

	newer := logAt(t, 1, base.Add(2*time.Second))
	store.AddLog(&newer)
	test.Render(t, page.view, screen)
	row, _ := page.table.table.GetSelection()
	assert.Equal(t, 1, row)

	older := logAt(t, 2, base.Add(time.Second))
	store.AddLog(&older)
	test.Render(t, page.view, screen)

	row, _ = page.table.table.GetSelection()
	// Each payload has two log records, so the selected first record of
	// the newer payload moves from row 1 to row 3.
	assert.Equal(t, 3, row)
	assert.Equal(t, newer.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0).TraceID(), store.GetFilteredLogByIdx(row-1).Log.TraceID())
}
