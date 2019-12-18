package metric

import (
	"github.com/prometheus/client_golang/prometheus"
)

func RegisterMetrics() (err error) {
	// Metrics
	prometheus.MustRegister(Counter)
	prometheus.MustRegister(Gauge)
	prometheus.MustRegister(Histogram)
	prometheus.MustRegister(Summary)

	// Labeled metrics
	prometheus.MustRegister(LabeledCounter)
	prometheus.MustRegister(LabeledGauge)
	prometheus.MustRegister(LabeledHistogram)
	prometheus.MustRegister(LabeledSummary)
	return
}
