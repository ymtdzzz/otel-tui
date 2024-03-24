package telemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateCache(t *testing.T) {
	t.Skip("TODO")
}

func TestGetSpansByTraceID(t *testing.T) {
	c := NewTraceCache()
	spans := []*SpanData{{}}
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
	t.Skip("TODO")
}

func TestGetSpanByID(t *testing.T) {
	t.Skip("TODO")
}
