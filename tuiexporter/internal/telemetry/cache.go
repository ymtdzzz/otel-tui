package telemetry

import (
	"context"
)

// SpanDataMap is a map of span id to span data
// This is used to quickly look up a span by its id
type SpanDataMap map[string]*SpanData

// TraceSpanDataMap is a map of trace id to a slice of spans
// This is used to quickly look up all spans in a trace
type TraceSpanDataMap map[string][]*SpanData

// TraceServiceSpanDataMap is a map of trace id and service name to a slice of spans
// This is used to quickly look up all spans in a trace for a service
type TraceServiceSpanDataMap map[string]map[string][]*SpanData

// TraceCache is a cache of trace spans
type TraceCache struct {
	spanid2span    SpanDataMap
	traceid2spans  TraceSpanDataMap
	tracesvc2spans TraceServiceSpanDataMap
}

// NewTraceCache returns a new trace cache
func NewTraceCache() *TraceCache {
	return &TraceCache{
		spanid2span:    SpanDataMap{},
		traceid2spans:  TraceSpanDataMap{},
		tracesvc2spans: TraceServiceSpanDataMap{},
	}
}

// UpdateCache updates the cache with a new span
func (c *TraceCache) UpdateCache(sname string, data *SpanData) (newtracesvc bool) {
	c.spanid2span[data.Span.SpanID().String()] = data
	traceID := data.Span.TraceID().String()
	if ts, ok := c.traceid2spans[traceID]; ok {
		c.traceid2spans[traceID] = append(ts, data)
		if _, ok := c.tracesvc2spans[traceID][sname]; ok {
			c.tracesvc2spans[traceID][sname] = append(c.tracesvc2spans[traceID][sname], data)
		} else {
			c.tracesvc2spans[traceID][sname] = []*SpanData{data}
			newtracesvc = true
		}
	} else {
		c.traceid2spans[traceID] = []*SpanData{data}
		c.tracesvc2spans[traceID] = map[string][]*SpanData{sname: {data}}
		newtracesvc = true
	}

	return newtracesvc
}

// DeleteCache deletes a list of spans from the cache
func (c *TraceCache) DeleteCache(serviceSpans []*SpanData) {
	// FIXME: more efficient way ?
	for _, ss := range serviceSpans {
		traceID := ss.Span.TraceID().String()
		sname, _ := ss.ResourceSpan.Resource().Attributes().Get("service.name")

		if spans, ok := c.GetSpansByTraceIDAndSvc(ss.Span.TraceID().String(), sname.AsString()); ok {
			for _, s := range spans {
				delete(c.spanid2span, s.Span.SpanID().String())
			}
		}
		delete(c.tracesvc2spans[traceID], sname.AsString())
		if len(c.tracesvc2spans[traceID]) == 0 {
			delete(c.tracesvc2spans, traceID)
			// delete spans in traceid2spans only if there are no spans left in tracesvc2spans
			// for better performance
			delete(c.traceid2spans, traceID)
		}
	}
}

// GetSpansByTraceID returns all spans for a given trace id
func (c *TraceCache) GetSpansByTraceID(traceID string) ([]*SpanData, bool) {
	spans, ok := c.traceid2spans[traceID]
	return spans, ok
}

// GetSpansByTraceIDAndSvc returns all spans for a given trace id and service name
func (c *TraceCache) GetSpansByTraceIDAndSvc(traceID, svc string) ([]*SpanData, bool) {
	if spans, ok := c.tracesvc2spans[traceID]; ok {
		if ss, ok := spans[svc]; ok {
			return ss, ok
		}
	}
	return nil, false
}

// GetSpanByID returns a span by its id
func (c *TraceCache) GetSpanByID(spanID string) (*SpanData, bool) {
	span, ok := c.spanid2span[spanID]
	return span, ok
}

func (c *TraceCache) GetRootSpanByID(ctx context.Context, spanID string) (*SpanData, bool) {
	select {
	case <-ctx.Done():
		// context canceled
		return nil, false
	default:
	}
	span, ok := c.GetSpanByID(spanID)
	if !ok {
		// invalid span ID or orphan span
		// TODO: orphan span should be marked to notify the users of that
		return nil, false
	}
	if !span.Span.ParentSpanID().IsEmpty() {
		return c.GetRootSpanByID(ctx, span.Span.ParentSpanID().String())
	}
	// This is the root span
	return span, ok
}

func (c *TraceCache) flush() {
	c.spanid2span = SpanDataMap{}
	c.traceid2spans = TraceSpanDataMap{}
	c.tracesvc2spans = TraceServiceSpanDataMap{}
}

// TraceLogDataMap is a map of trace id to a slice of logs
// This is used to quickly look up all logs in a trace
type TraceLogDataMap map[string][]*LogData

// LogCache is a cache of logs
type LogCache struct {
	traceid2logs TraceLogDataMap
}

// NewLogCache returns a new log cache
func NewLogCache() *LogCache {
	return &LogCache{
		traceid2logs: TraceLogDataMap{},
	}
}

// UpdateCache updates the cache with a new log
func (c *LogCache) UpdateCache(data *LogData) {
	traceID := data.Log.TraceID().String()
	if ts, ok := c.traceid2logs[traceID]; ok {
		c.traceid2logs[traceID] = append(ts, data)
	} else {
		c.traceid2logs[traceID] = []*LogData{data}
	}
}

// DeleteCache deletes a list of logs from the cache
func (c *LogCache) DeleteCache(logs []*LogData) {
	for _, l := range logs {
		traceID := l.Log.TraceID().String()
		if _, ok := c.traceid2logs[traceID]; ok {
			for i, log := range c.traceid2logs[traceID] {
				if log == l {
					c.traceid2logs[traceID] = append(c.traceid2logs[traceID][:i], c.traceid2logs[traceID][i+1:]...)
					break
				}
			}
		}
	}
}

// GetLogsByTraceID returns all logs for a given trace id
func (c *LogCache) GetLogsByTraceID(traceID string) ([]*LogData, bool) {
	logs, ok := c.traceid2logs[traceID]
	return logs, ok
}

func (c *LogCache) flush() {
	c.traceid2logs = TraceLogDataMap{}
}

// MetricServiceMetricDataMap is a map of service name and metric name to a slice of metrics
// This is used to quickly look up datapoints in a service metric
type MetricServiceMetricDataMap map[string]map[string][]*MetricData

// MetricCache is a cache of metrics
type MetricCache struct {
	svcmetric2metrics MetricServiceMetricDataMap
}

// NewMetricCache returns a new metric cache
func NewMetricCache() *MetricCache {
	return &MetricCache{
		svcmetric2metrics: MetricServiceMetricDataMap{},
	}
}

// UpdateCache updates the cache with a new metric
func (c *MetricCache) UpdateCache(sname string, data *MetricData) {
	mname := data.Metric.Name()
	if sms, ok := c.svcmetric2metrics[sname]; ok {
		if _, ok := sms[mname]; ok {
			c.svcmetric2metrics[sname][mname] = append(c.svcmetric2metrics[sname][mname], data)
		} else {
			c.svcmetric2metrics[sname][mname] = []*MetricData{data}
		}
	} else {
		c.svcmetric2metrics[sname] = map[string][]*MetricData{mname: {data}}
	}
}

// DeleteCache deletes a list of metrics from the cache
func (c *MetricCache) DeleteCache(metrics []*MetricData) {
	for _, m := range metrics {
		sname := "N/A"
		if snameattr, ok := m.ResourceMetric.Resource().Attributes().Get("service.name"); ok {
			sname = snameattr.AsString()
		}
		mname := m.Metric.Name()
		if _, ok := c.svcmetric2metrics[sname][mname]; ok {
			for i, metric := range c.svcmetric2metrics[sname][mname] {
				if metric == m {
					c.svcmetric2metrics[sname][mname] = append(c.svcmetric2metrics[sname][mname][:i], c.svcmetric2metrics[sname][mname][i+1:]...)
					if len(c.svcmetric2metrics[sname][mname]) == 0 {
						delete(c.svcmetric2metrics[sname], mname)
						if len(c.svcmetric2metrics[sname]) == 0 {
							delete(c.svcmetric2metrics, sname)
						}
					}
				}
			}
		}
	}
}

// GetMetricsBySvcAndMetricName returns all metrics for a given service name and metric name
func (c *MetricCache) GetMetricsBySvcAndMetricName(sname, mname string) ([]*MetricData, bool) {
	if sms, ok := c.svcmetric2metrics[sname]; ok {
		if ms, ok := sms[mname]; ok {
			return ms, ok
		}
	}
	return nil, false
}

func (c *MetricCache) flush() {
	c.svcmetric2metrics = MetricServiceMetricDataMap{}
}
