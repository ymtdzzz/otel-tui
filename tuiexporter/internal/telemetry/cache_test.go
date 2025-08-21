package telemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pmetric"
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

func TestMetricCacheAggregateView(t *testing.T) {
	cache := NewMetricCache()

	// Create test metric data
	// Note: This is a simplified test - in reality, you'd create proper pmetric.Metric objects
	// For testing purposes, we'll verify the cache structure and basic functionality

	t.Run("GetAllMetricNames_empty_cache", func(t *testing.T) {
		names := cache.GetAllMetricNames()
		assert.Empty(t, names)
	})

	t.Run("GetMetricDataByName_nonexistent", func(t *testing.T) {
		data, metricType, exists := cache.GetMetricDataByName("nonexistent")
		assert.Nil(t, data)
		assert.Equal(t, pmetric.MetricTypeEmpty, metricType)
		assert.False(t, exists)
	})

	t.Run("GetMetricNamesSummary_empty_cache", func(t *testing.T) {
		summaries := cache.GetMetricNamesSummary()
		assert.Empty(t, summaries)
	})

	// Test cache initialization
	t.Run("cache_initialization", func(t *testing.T) {
		assert.NotNil(t, cache.svcmetric2metrics)
		assert.NotNil(t, cache.name2metrics)
		assert.NotNil(t, cache.name2types)
	})
}
