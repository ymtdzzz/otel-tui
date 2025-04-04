package app

import (
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

type PushTracesMsg struct {
	Traces *ptrace.Traces
}

type PushMetricsMsg struct {
	Metrics *pmetric.Metrics
}

type PushLogsMsg struct {
	Logs *plog.Logs
}

type rotateTabMsg struct{}
