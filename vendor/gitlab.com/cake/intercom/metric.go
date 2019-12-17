package intercom

import (
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	prometheus.MustRegister(upstreamCounter)
	prometheus.MustRegister(upstreamDuration)
}

const (
	metricNs          = "intercom"
	subSystemUpstream = "upstream"

	labelHost = "host"
	labelCode = "code"
)

var (
	upstreamCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNs,
			Subsystem: subSystemUpstream,
			Name:      "upstream_sent_total",
			Help:      "Total upstream request number sent",
		},
		[]string{labelHost, labelCode},
	)

	upstreamDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricNs,
			Subsystem: subSystemUpstream,
			Name:      "upstream_duration_seconds",
			Help:      "Total upstream request duration",
			Buckets:   prometheus.ExponentialBuckets(0.1, 2, 10),
		},
		[]string{labelHost, labelCode},
	)
)
