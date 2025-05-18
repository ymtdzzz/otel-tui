package test

import (
	"fmt"
	"testing"

	"go.opentelemetry.io/collector/pdata/pmetric"
)

type GeneratedMetrics struct {
	Metrics  []*pmetric.Metric
	RMetrics []*pmetric.ResourceMetrics
	SMetrics []*pmetric.ScopeMetrics
}

func GenerateOTLPGaugeMetricsPayload(t *testing.T, resourceCount int, scopeCount []int, dpCount [][]int) (pmetric.Metrics, *GeneratedMetrics) {
	t.Helper()

	generatedMetrics := &GeneratedMetrics{
		Metrics:  []*pmetric.Metric{},
		RMetrics: []*pmetric.ResourceMetrics{},
		SMetrics: []*pmetric.ScopeMetrics{},
	}
	metricData := pmetric.NewMetrics()

	// Create and populate resource data
	metricData.ResourceMetrics().EnsureCapacity(resourceCount)
	for resourceIndex := 0; resourceIndex < resourceCount; resourceIndex++ {
		scopeCount := scopeCount[resourceIndex]
		resourceMetric := metricData.ResourceMetrics().AppendEmpty()
		fillResource(t, resourceMetric.Resource(), resourceIndex)
		generatedMetrics.RMetrics = append(generatedMetrics.RMetrics, &resourceMetric)

		// Create and populate instrumentation scope data
		resourceMetric.ScopeMetrics().EnsureCapacity(scopeCount)
		for scopeIndex := 0; scopeIndex < scopeCount; scopeIndex++ {
			scopeMetric := resourceMetric.ScopeMetrics().AppendEmpty()
			fillScope(t, scopeMetric.Scope(), resourceIndex, scopeIndex)
			generatedMetrics.SMetrics = append(generatedMetrics.SMetrics, &scopeMetric)

			// Create and populate metrics
			// 1 metric per scope
			scopeMetric.Metrics().EnsureCapacity(1)
			metric := scopeMetric.Metrics().AppendEmpty()
			fillMetric(t, metric, resourceIndex, scopeIndex)
			gauge := metric.SetEmptyGauge()
			gauge.DataPoints().EnsureCapacity(dpCount[resourceIndex][scopeIndex])
			for dpIndex := 0; dpIndex < dpCount[resourceIndex][scopeIndex]; dpIndex++ {
				dp := metric.Gauge().DataPoints().AppendEmpty()
				fillNumberDataPoint(t, dp, dpIndex)
			}
			generatedMetrics.Metrics = append(generatedMetrics.Metrics, &metric)
		}
	}

	return metricData, generatedMetrics
}

func GenerateOTLPHistogramMetricsPayload(t *testing.T, resourceCount int, scopeCount []int, dpCount [][]int) (pmetric.Metrics, *GeneratedMetrics) {
	t.Helper()

	generatedMetrics := &GeneratedMetrics{
		Metrics:  []*pmetric.Metric{},
		RMetrics: []*pmetric.ResourceMetrics{},
		SMetrics: []*pmetric.ScopeMetrics{},
	}
	metricData := pmetric.NewMetrics()

	// Create and populate resource data
	metricData.ResourceMetrics().EnsureCapacity(resourceCount)
	for resourceIndex := 0; resourceIndex < resourceCount; resourceIndex++ {
		scopeCount := scopeCount[resourceIndex]
		resourceMetric := metricData.ResourceMetrics().AppendEmpty()
		fillResource(t, resourceMetric.Resource(), resourceIndex)
		generatedMetrics.RMetrics = append(generatedMetrics.RMetrics, &resourceMetric)

		// Create and populate instrumentation scope data
		resourceMetric.ScopeMetrics().EnsureCapacity(scopeCount)
		for scopeIndex := 0; scopeIndex < scopeCount; scopeIndex++ {
			scopeMetric := resourceMetric.ScopeMetrics().AppendEmpty()
			fillScope(t, scopeMetric.Scope(), resourceIndex, scopeIndex)
			generatedMetrics.SMetrics = append(generatedMetrics.SMetrics, &scopeMetric)

			// Create and populate metrics
			// 1 metric per scope
			scopeMetric.Metrics().EnsureCapacity(1)
			metric := scopeMetric.Metrics().AppendEmpty()
			fillMetric(t, metric, resourceIndex, scopeIndex)
			histogram := metric.SetEmptyHistogram()
			histogram.DataPoints().EnsureCapacity(dpCount[resourceIndex][scopeIndex])
			for dpIndex := 0; dpIndex < dpCount[resourceIndex][scopeIndex]; dpIndex++ {
				dp := metric.Histogram().DataPoints().AppendEmpty()
				fillHistogramDataPoint(t, dp, dpIndex)
			}
			generatedMetrics.Metrics = append(generatedMetrics.Metrics, &metric)
		}
	}

	return metricData, generatedMetrics
}

func fillMetric(t *testing.T, m pmetric.Metric, resourceIndex, scopeIndex int) {
	t.Helper()

	m.SetName(fmt.Sprintf("metric %d-%d", resourceIndex, scopeIndex))
	m.SetUnit("test unit")
	m.SetDescription("test description")
}

func fillNumberDataPoint(t *testing.T, dp pmetric.NumberDataPoint, dpIndex int) {
	t.Helper()

	dp.SetDoubleValue(float64(dpIndex + 1))
	dp.SetFlags(pmetric.DefaultDataPointFlags.WithNoRecordedValue(false))
	// TODO: examplers
	dp.Attributes().PutInt("dp index", int64(dpIndex))
}

func fillHistogramDataPoint(t *testing.T, dp pmetric.HistogramDataPoint, dpIndex int) {
	t.Helper()

	if dpIndex < 0 {
		dp.SetCount(0)
	} else {
		dp.SetCount(uint64(dpIndex + 1)) // #nosec G115
	}
	dp.SetSum(float64(dpIndex + 1))
	dp.BucketCounts().FromRaw([]uint64{0, 10, 20, 5, 0})
	dp.ExplicitBounds().FromRaw([]float64{0, 10, 20, 30})
	dp.SetFlags(pmetric.DefaultDataPointFlags.WithNoRecordedValue(false))
	// TODO: examplers
	dp.Attributes().PutInt("dp index", int64(dpIndex))
}
