package telemetry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/test"
)

func TestSortType(t *testing.T) {
	tests := []struct {
		name            string
		input           SortType
		wantIsNone      bool
		wantIsDesc      bool
		wantHeaderLabel string
	}{
		{
			name:            "SORT_TYPE_NONE",
			input:           SORT_TYPE_NONE,
			wantIsNone:      true,
			wantIsDesc:      false,
			wantHeaderLabel: "N/A",
		},
		{
			name:            "SORT_TYPE_LATENCY_DESC",
			input:           SORT_TYPE_LATENCY_DESC,
			wantIsNone:      false,
			wantIsDesc:      true,
			wantHeaderLabel: "Latency",
		},
		{
			name:            "SORT_TYPE_LATENCY_ASC",
			input:           SORT_TYPE_LATENCY_ASC,
			wantIsNone:      false,
			wantIsDesc:      false,
			wantHeaderLabel: "Latency",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantIsNone, tt.input.IsNone())
			assert.Equal(t, tt.wantIsDesc, tt.input.IsDesc())
			assert.Equal(t, tt.wantHeaderLabel, tt.input.GetHeaderLabel())
		})
	}
}

func TestSortSvcSpans(t *testing.T) {
	baseSvcSpans := SvcSpans{
		&SpanData{
			Span: test.GenerateSpanWithDuration(t, "100ms", 100*time.Millisecond),
		},
		&SpanData{
			Span: test.GenerateSpanWithDuration(t, "75µs", 50*time.Microsecond),
		},
		&SpanData{
			Span: test.GenerateSpanWithDuration(t, "230ms", 230*time.Millisecond),
		},
		&SpanData{
			Span: test.GenerateSpanWithDuration(t, "101ms", 101*time.Millisecond),
		},
		&SpanData{
			Span: test.GenerateSpanWithDuration(t, "50ns", 50*time.Nanosecond),
		},
	}

	tests := []struct {
		name     string
		sortType SortType
		input    SvcSpans
		want     SvcSpans
	}{
		{
			name:     "SORT_TYPE_NONE",
			sortType: SORT_TYPE_NONE,
			input:    append(SvcSpans{}, baseSvcSpans...),
			want: SvcSpans{
				&SpanData{
					Span: test.GenerateSpanWithDuration(t, "100ms", 100*time.Millisecond),
				},
				&SpanData{
					Span: test.GenerateSpanWithDuration(t, "75µs", 50*time.Microsecond),
				},
				&SpanData{
					Span: test.GenerateSpanWithDuration(t, "230ms", 230*time.Millisecond),
				},
				&SpanData{
					Span: test.GenerateSpanWithDuration(t, "101ms", 101*time.Millisecond),
				},
				&SpanData{
					Span: test.GenerateSpanWithDuration(t, "50ns", 50*time.Nanosecond),
				},
			},
		},
		{
			name:     "SORT_TYPE_LATENCY_DESC",
			sortType: SORT_TYPE_LATENCY_DESC,
			input:    append(SvcSpans{}, baseSvcSpans...),
			want: SvcSpans{
				&SpanData{
					Span: test.GenerateSpanWithDuration(t, "230ms", 230*time.Millisecond),
				},
				&SpanData{
					Span: test.GenerateSpanWithDuration(t, "101ms", 101*time.Millisecond),
				},
				&SpanData{
					Span: test.GenerateSpanWithDuration(t, "100ms", 100*time.Millisecond),
				},
				&SpanData{
					Span: test.GenerateSpanWithDuration(t, "75µs", 50*time.Microsecond),
				},
				&SpanData{
					Span: test.GenerateSpanWithDuration(t, "50ns", 50*time.Nanosecond),
				},
			},
		},
		{
			name:     "SORT_TYPE_LATENCY_ASC",
			sortType: SORT_TYPE_LATENCY_ASC,
			input:    append(SvcSpans{}, baseSvcSpans...),
			want: SvcSpans{
				&SpanData{
					Span: test.GenerateSpanWithDuration(t, "50ns", 50*time.Nanosecond),
				},
				&SpanData{
					Span: test.GenerateSpanWithDuration(t, "75µs", 50*time.Microsecond),
				},
				&SpanData{
					Span: test.GenerateSpanWithDuration(t, "100ms", 100*time.Millisecond),
				},
				&SpanData{
					Span: test.GenerateSpanWithDuration(t, "101ms", 101*time.Millisecond),
				},
				&SpanData{
					Span: test.GenerateSpanWithDuration(t, "230ms", 230*time.Millisecond),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sortSvcSpans(tt.input, tt.sortType)
			assert.Equal(t, tt.want, tt.input)
		})
	}
}
