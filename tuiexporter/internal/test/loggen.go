package test

import (
	"fmt"
	"testing"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
)

var (
	logTimestamp         = pcommon.NewTimestampFromTime(time.Date(2022, 10, 21, 7, 10, 2, 100000000, time.UTC))
	logObservedTimestamp = pcommon.NewTimestampFromTime(time.Date(2022, 10, 21, 7, 10, 2, 200000000, time.UTC))
)

type GeneratedLogs struct {
	Logs  []*plog.LogRecord
	RLogs []*plog.ResourceLogs
	SLogs []*plog.ScopeLogs
}

func GenerateOTLPLogsPayload(t *testing.T, traceID, resourceCount int, scopeCount []int, spanCount [][]int) (plog.Logs, *GeneratedLogs) {
	t.Helper()

	generatedLogs := &GeneratedLogs{
		Logs:  []*plog.LogRecord{},
		RLogs: []*plog.ResourceLogs{},
		SLogs: []*plog.ScopeLogs{},
	}
	logData := plog.NewLogs()
	uniqueSpanIndex := 0

	// Create and populate resource data
	logData.ResourceLogs().EnsureCapacity(resourceCount)
	for resourceIndex := 0; resourceIndex < resourceCount; resourceIndex++ {
		scopeCount := scopeCount[resourceIndex]
		resourceLog := logData.ResourceLogs().AppendEmpty()
		fillResource(t, resourceLog.Resource(), resourceIndex)
		generatedLogs.RLogs = append(generatedLogs.RLogs, &resourceLog)

		// Create and populate instrumentation scope data
		resourceLog.ScopeLogs().EnsureCapacity(scopeCount)
		for scopeIndex := 0; scopeIndex < scopeCount; scopeIndex++ {
			spanCount := spanCount[resourceIndex][scopeIndex]
			scopeLog := resourceLog.ScopeLogs().AppendEmpty()
			fillScope(t, scopeLog.Scope(), resourceIndex, scopeIndex)
			generatedLogs.SLogs = append(generatedLogs.SLogs, &scopeLog)

			//Create and populate spans
			scopeLog.LogRecords().EnsureCapacity(spanCount)
			for spanIndex := 0; spanIndex < spanCount; spanIndex++ {
				// 2 logs per span
				record1 := scopeLog.LogRecords().AppendEmpty()
				fillLog(t, record1, traceID, resourceIndex, scopeIndex, spanIndex, 0, uniqueSpanIndex)
				record2 := scopeLog.LogRecords().AppendEmpty()
				fillLog(t, record2, traceID, resourceIndex, scopeIndex, spanIndex, 1, uniqueSpanIndex)
				generatedLogs.Logs = append(generatedLogs.Logs, &record1, &record2)
				uniqueSpanIndex++
			}
		}
	}

	return logData, generatedLogs
}

func fillLog(t *testing.T, l plog.LogRecord, traceID, resourceIndex, scopeIndex, spanIndex, logIndex, uniqueSpanIndex int) {
	t.Helper()
	spanID := [8]byte{byte(uniqueSpanIndex + 1)}

	l.SetTraceID([16]byte{byte(traceID)})
	l.SetSpanID(spanID)

	l.Body().SetStr(fmt.Sprintf("log body %d-%d-%d-%d", resourceIndex, scopeIndex, spanIndex, logIndex))

	l.SetSeverityNumber(plog.SeverityNumberInfo)
	l.SetSeverityText("INFO")

	l.SetTimestamp(logTimestamp)
	l.SetObservedTimestamp(logObservedTimestamp)

	l.SetDroppedAttributesCount(3)
	l.Attributes().PutInt("span index", int64(spanIndex))
	l.SetDroppedAttributesCount(3)
}
