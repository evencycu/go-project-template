package ginmetrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	strAssert = "find unmatched metrics type and opts when metric %s registers with type %s"

	KeyReqSz     = "reqSz"
	KeyResSz     = "resSz"
	KeyReqDur    = "reqDur"
	KeyReqCnt    = "reqCnt"
	KeyReqDurAll = "reqDurAll"

	TypeCounter      = "counter"
	TypeCounterVec   = "counterVec"
	TypeGauge        = "gauge"
	TypeGaugeVec     = "gaugeVec"
	TypeHistogram    = "histogram"
	TypeHistogramVec = "histogramVec"
	TypeSummary      = "summary"
	TypeSummaryVec   = "summaryVec"

	defaultMetricPath = "/metrics"
)

var (
	// 0.1 0.2 0.4 0.8 1.6 3.2 6.4
	DefaultDurationBucket = prometheus.ExponentialBuckets(0.1, 2, 7)

	// 100 200 400 800 1600
	DefaultSizeBucket = prometheus.ExponentialBuckets(100, 2, 5)
)
