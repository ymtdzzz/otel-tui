package telemetry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

func TestGetSpansByTraceID(t *testing.T) {
	c := NewTraceCache()
	spans := []*SpanData{}
	c.traceid2spans["traceid"] = spans

	tests := []struct {
		name     string
		traceID  string
		wantdata []*SpanData
		wantok   bool
	}{
		{
			name:     "traceid exists",
			traceID:  "traceid",
			wantdata: spans,
			wantok:   true,
		},
		{
			name:     "traceid does not exist",
			traceID:  "traceid2",
			wantdata: nil,
			wantok:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotdata, gotok := c.GetSpansByTraceID(tt.traceID)
			assert.Equal(t, tt.wantdata, gotdata)
			assert.Equal(t, tt.wantok, gotok)
		})
	}
}

func TestGetSpansByTraceIDAndSvc(t *testing.T) {
	c := NewTraceCache()
	spans := []*SpanData{}
	c.tracesvc2spans["traceid"] = map[string][]*SpanData{"svc-name": spans}

	tests := []struct {
		name     string
		traceID  string
		svcName  string
		wantdata []*SpanData
		wantok   bool
	}{
		{
			name:     "traceid and service exists",
			traceID:  "traceid",
			svcName:  "svc-name",
			wantdata: spans,
			wantok:   true,
		},
		{
			name:     "traceid exists but service does not",
			traceID:  "traceid",
			svcName:  "non-existent-service",
			wantdata: nil,
			wantok:   false,
		},
		{
			name:     "traceid does not exist",
			traceID:  "non-existent-traceid",
			svcName:  "svc-name",
			wantdata: nil,
			wantok:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotdata, gotok := c.GetSpansByTraceIDAndSvc(tt.traceID, tt.svcName)
			assert.Equal(t, tt.wantdata, gotdata)
			assert.Equal(t, tt.wantok, gotok)
		})
	}
}

func TestGetSpanByID(t *testing.T) {
	c := NewTraceCache()
	span := &SpanData{}
	c.spanid2span["spanid"] = span

	tests := []struct {
		name     string
		spanID   string
		wantdata *SpanData
		wantok   bool
	}{
		{
			name:     "spanid exists",
			spanID:   "spanid",
			wantdata: span,
			wantok:   true,
		},
		{
			name:     "spanid does not exist",
			spanID:   "non-existent-spanid",
			wantdata: nil,
			wantok:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotdata, gotok := c.GetSpanByID(tt.spanID)
			assert.Equal(t, tt.wantdata, gotdata)
			assert.Equal(t, tt.wantok, gotok)
		})
	}
}

func TestGetRootSpanByID(t *testing.T) {
	c := NewTraceCache()
	var (
		spanIDA            pcommon.SpanID = [8]byte{byte(1)}
		spanIDB            pcommon.SpanID = [8]byte{byte(2)}
		spanIDC            pcommon.SpanID = [8]byte{byte(3)}
		orphanParentSpanID pcommon.SpanID = [8]byte{byte(4)}
		orphanChildSpanID  pcommon.SpanID = [8]byte{byte(5)}
	)
	spanA := ptrace.NewSpan()
	spanA.SetParentSpanID(pcommon.NewSpanIDEmpty()) // root
	spanB := ptrace.NewSpan()
	spanB.SetParentSpanID(spanIDA) // A's child
	spanC := ptrace.NewSpan()
	spanC.SetParentSpanID(spanIDB) // B's child
	spanOrphanChild := ptrace.NewSpan()
	spanOrphanChild.SetParentSpanID(orphanParentSpanID) // non-existent parent's child
	spanDataA := &SpanData{
		Span: &spanA,
	}
	spanDataB := &SpanData{
		Span: &spanB,
	}
	spanDataC := &SpanData{
		Span: &spanC,
	}
	orphanChildSpanData := &SpanData{
		Span: &spanOrphanChild,
	}
	c.spanid2span[spanIDA.String()] = spanDataA
	c.spanid2span[spanIDB.String()] = spanDataB
	c.spanid2span[spanIDC.String()] = spanDataC
	c.spanid2span[orphanChildSpanID.String()] = orphanChildSpanData

	tests := []struct {
		name     string
		spanID   string
		wantdata *SpanData
		wantok   bool
	}{
		{
			"root span exists from C",
			spanIDC.String(),
			spanDataA,
			true,
		},
		{
			"root span exists from B",
			spanIDB.String(),
			spanDataA,
			true,
		},
		{
			"root span exists from A itself",
			spanIDA.String(),
			spanDataA,
			true,
		},
		{
			"orphan span",
			orphanChildSpanID.String(),
			nil,
			false,
		},
		{
			name:     "spanid does not exist",
			spanID:   "non-existent-spanid",
			wantdata: nil,
			wantok:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotdata, gotok := c.GetRootSpanByID(context.Background(), tt.spanID)
			assert.Equal(t, tt.wantdata, gotdata)
			assert.Equal(t, tt.wantok, gotok)
		})
	}
}

func TestGetLogsByTraceID(t *testing.T) {
	c := NewLogCache()
	logs := []*LogData{}
	c.traceid2logs["traceid"] = logs

	tests := []struct {
		name     string
		traceID  string
		wantdata []*LogData
		wantok   bool
	}{
		{
			name:     "traceid exists",
			traceID:  "traceid",
			wantdata: logs,
			wantok:   true,
		},
		{
			name:     "traceid does not exist",
			traceID:  "traceid2",
			wantdata: nil,
			wantok:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotdata, gotok := c.GetLogsByTraceID(tt.traceID)
			assert.Equal(t, tt.wantdata, gotdata)
			assert.Equal(t, tt.wantok, gotok)
		})
	}
}

func TestGetMetricsBySvcAndMetricName(t *testing.T) {
	c := NewMetricCache()
	metrics := []*MetricData{}
	c.svcmetric2metrics["sname"] = map[string][]*MetricData{"mname": metrics}

	tests := []struct {
		name     string
		sname    string
		mname    string
		wantdata []*MetricData
		wantok   bool
	}{
		{
			name:     "service and metrics exists",
			sname:    "sname",
			mname:    "mname",
			wantdata: metrics,
			wantok:   true,
		},
		{
			name:     "service exists but metrics does not",
			sname:    "sname",
			mname:    "non-existent-metric",
			wantdata: nil,
			wantok:   false,
		},
		{
			name:     "service does not exist",
			sname:    "non-existent-sname",
			mname:    "mname",
			wantdata: nil,
			wantok:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotdata, gotok := c.GetMetricsBySvcAndMetricName(tt.sname, tt.mname)
			assert.Equal(t, tt.wantdata, gotdata)
			assert.Equal(t, tt.wantok, gotok)
		})
	}
}
