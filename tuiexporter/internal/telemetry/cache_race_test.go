package telemetry

import (
	"fmt"
	"sync"
	"testing"

	"github.com/jonboulle/clockwork"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/test"
)

func TestTraceCache_ConcurrentMapAccess(t *testing.T) {
	store := NewStore(clockwork.NewRealClock())
	cache := store.GetTraceCache()

	const (
		writers         = 4
		writesPerWorker = 250
		traceIDSpace    = 200 // traceID is encoded as a single byte by the test generator
	)

	traceIDs := make([]string, traceIDSpace)
	for i := range traceIDs {
		traceIDs[i] = fmt.Sprintf("%032x", byte(i+1))
	}

	var writerWG, readerWG sync.WaitGroup
	stop := make(chan struct{})

	// Reader goroutine: mimics the tview draw loop hammering the cache.
	readerWG.Add(1)
	go func() {
		defer readerWG.Done()
		for {
			select {
			case <-stop:
				return
			default:
			}
			for _, traceID := range traceIDs {
				cache.HasErrorByTraceIDAndSvc(traceID, "test-service-1")
				cache.GetSpansByTraceID(traceID)
			}
		}
	}()

	// Writer goroutines: simulate multiple OTLP receivers ingesting spans.
	for w := 0; w < writers; w++ {
		writerWG.Add(1)
		go func(workerID int) {
			defer writerWG.Done()
			for i := 0; i < writesPerWorker; i++ {
				traceID := ((workerID*writesPerWorker + i) % traceIDSpace) + 1
				payload, _ := test.GenerateOTLPTracesPayload(t, traceID, 1, []int{1}, [][]int{{2}})
				store.AddSpan(&payload)
			}
		}(w)
	}

	writerWG.Wait()
	close(stop)
	readerWG.Wait()
}

func TestLogCache_ConcurrentMapAccess(t *testing.T) {
	store := NewStore(clockwork.NewRealClock())
	cache := store.GetLogCache()

	const (
		writers         = 4
		writesPerWorker = 250
		traceIDSpace    = 200 // traceID is encoded as a single byte by the test generator
	)

	traceIDs := make([]string, traceIDSpace)
	for i := range traceIDs {
		traceIDs[i] = fmt.Sprintf("%032x", byte(i+1))
	}

	var writerWG, readerWG sync.WaitGroup
	stop := make(chan struct{})

	readerWG.Add(1)
	go func() {
		defer readerWG.Done()
		for {
			select {
			case <-stop:
				return
			default:
			}
			for _, traceID := range traceIDs {
				cache.GetLogsByTraceID(traceID)
			}
		}
	}()

	for w := 0; w < writers; w++ {
		writerWG.Add(1)
		go func(workerID int) {
			defer writerWG.Done()
			for i := 0; i < writesPerWorker; i++ {
				traceID := ((workerID*writesPerWorker + i) % traceIDSpace) + 1
				payload, _ := test.GenerateOTLPLogsPayload(t, traceID, 1, []int{1}, [][]int{{2}})
				store.AddLog(&payload)
			}
		}(w)
	}

	writerWG.Wait()
	close(stop)
	readerWG.Wait()
}

func TestMetricCache_ConcurrentMapAccess(t *testing.T) {
	store := NewStore(clockwork.NewRealClock())
	cache := store.GetMetricCache()

	const (
		writers         = 4
		writesPerWorker = 250
		serviceCount    = 8
		scopesPerSvc    = 4
	)

	type svcMetric struct{ svc, metric string }
	pairs := make([]svcMetric, 0, serviceCount*scopesPerSvc)
	for r := 0; r < serviceCount; r++ {
		for s := 0; s < scopesPerSvc; s++ {
			pairs = append(pairs, svcMetric{
				svc:    fmt.Sprintf("test-service-%d", r+1),
				metric: fmt.Sprintf("metric %d-%d", r, s),
			})
		}
	}

	var writerWG, readerWG sync.WaitGroup
	stop := make(chan struct{})

	readerWG.Add(1)
	go func() {
		defer readerWG.Done()
		for {
			select {
			case <-stop:
				return
			default:
			}
			for _, p := range pairs {
				cache.GetMetricsBySvcAndMetricName(p.svc, p.metric)
			}
		}
	}()

	scopeCounts := make([]int, serviceCount)
	dpCounts := make([][]int, serviceCount)
	for r := 0; r < serviceCount; r++ {
		scopeCounts[r] = scopesPerSvc
		dpCounts[r] = make([]int, scopesPerSvc)
		for s := 0; s < scopesPerSvc; s++ {
			dpCounts[r][s] = 1
		}
	}

	for w := 0; w < writers; w++ {
		writerWG.Add(1)
		go func() {
			defer writerWG.Done()
			for i := 0; i < writesPerWorker; i++ {
				payload, _ := test.GenerateOTLPGaugeMetricsPayload(t, serviceCount, scopeCounts, dpCounts)
				store.AddMetric(&payload)
			}
		}()
	}

	writerWG.Wait()
	close(stop)
	readerWG.Wait()
}
