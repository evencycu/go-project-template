package metric

import "github.com/prometheus/client_golang/prometheus"

const (
	namespace = "go"
	subsystem = "project_template"

	LabelService = "service"
	LabelType    = "type"
)

var (
	// 2 4 6 8 10
	DefaultBucket = prometheus.LinearBuckets(2, 2, 5)

	// 0.1 0.2 0.4 0.8 1.6 3.2 6.4 12.8 25.6
	DefaultObjectives = map[float64]float64{0.5: 0.05, 0.8: 0.05}
)

// Please check naming best practice in official document: https://prometheus.io/docs/practices/naming/

// metrics
var (
	Counter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "counter_total",
			Help:      "total counter.",
		},
	)

	Gauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "gauge_latest_value",
			Help:      "Number of latest updated value.",
		},
	)

	Histogram = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "histogram_data_value",
			Help:      "Histogram.",
			Buckets:   DefaultBucket,
		},
	)

	Summary = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace:  namespace,
			Subsystem:  subsystem,
			Name:       "summary_data_value",
			Help:       "Summary.",
			Objectives: DefaultObjectives,
		},
	)
)

// labeled metrics
var (
	LabeledCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "labeled_counter_total",
			Help:      "total counter.",
		},
		[]string{LabelService, LabelType},
	)

	LabeledGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "labeled_gauge_latest_value",
			Help:      "Number of latest updated value.",
		},
		[]string{LabelService, LabelType},
	)

	LabeledHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "labeled_histogram_data_value",
			Help:      "Histogram.",
			Buckets:   DefaultBucket,
		},
		[]string{LabelService, LabelType},
	)

	LabeledSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  namespace,
			Subsystem:  subsystem,
			Name:       "labeled_summary_data_value",
			Help:       "Summary.",
			Objectives: DefaultObjectives,
		},
		[]string{LabelService, LabelType},
	)
)
