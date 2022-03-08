package intercom

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gitlab.com/cake/gopkg"
)

func init() {
	prometheus.MustRegister(externalUpstreamCounter)
	prometheus.MustRegister(externalUpstreamDuration)
	prometheus.MustRegister(internalUpstreamCounter)
	prometheus.MustRegister(internalUpstreamDuration)
	prometheus.MustRegister(brokenPipeCounts)
	prometheus.MustRegister(proxyBrokenPipeCounts)
	prometheus.MustRegister(totalCrossRegionFailedRequestCounts)
	prometheus.MustRegister(totalCrossRegionRequestCounts)
	prometheus.MustRegister(currentCrossRegionRequestCounts)
	prometheus.MustRegister(crossRegionRequestDuration)
	prometheus.MustRegister(crossRegionRequestSize)
	prometheus.MustRegister(crossRegionResponseSize)
}

const (
	metricNs              = "intercom"
	subSystemUpstream     = "upstream"
	subSystemReverseProxy = "proxy"

	labelHost              = "host"
	labelHTTPCode          = "code"
	labelInternalCode      = "eCode"
	labelUpstream          = "upstream"
	labelUpstreamNamespace = "upstream_namespace"
	labelDownstream        = "downstream"
	labelFromNamespace     = "from_namespace"
	labelToNamespace       = "to_namespace"
)

var (
	externalUpstreamCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNs,
			Subsystem: subSystemUpstream,
			Name:      "external_upstream_sent_total",
			Help:      "Total external upstream request number sent",
		},
		[]string{
			labelHost,
			labelHTTPCode,
			labelInternalCode,
		},
	)

	externalUpstreamDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricNs,
			Subsystem: subSystemUpstream,
			Name:      "external_upstream_duration_seconds",
			Help:      "Total external upstream request duration",
			Buckets:   prometheus.ExponentialBuckets(0.1, 2, 10),
		},
		[]string{
			labelHost,
			labelHTTPCode,
			labelInternalCode,
		},
	)

	internalUpstreamCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNs,
			Subsystem: subSystemUpstream,
			Name:      "internal_upstream_sent_total",
			Help:      "Total internal upstream request number sent",
		},
		[]string{
			labelHost,
			labelInternalCode,
		},
	)

	internalUpstreamDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricNs,
			Subsystem: subSystemUpstream,
			Name:      "internal_upstream_duration_seconds",
			Help:      "Total internal upstream request duration",
			Buckets:   prometheus.ExponentialBuckets(0.1, 2, 10),
		},
		[]string{
			labelHost,
			labelInternalCode,
		},
	)

	brokenPipeCounts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNs,
			Subsystem: subSystemUpstream,
			Name:      "broken_pipes_counts",
			Help:      "Total count of broken pipes",
		},
		[]string{
			labelDownstream,
		},
	)

	proxyBrokenPipeCounts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNs,
			Subsystem: subSystemReverseProxy,
			Name:      "proxy_broken_pipes_counts",
			Help:      "Total count of reverse proxy broken pipes",
		},
		[]string{
			labelDownstream,
			labelUpstream,
			labelUpstreamNamespace,
		},
	)

	totalCrossRegionFailedRequestCounts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNs,
			Subsystem: subSystemReverseProxy,
			Name:      "cross_region_failed_requests_total",
			Help:      "Total count of cross region failed request",
		},
		[]string{
			labelFromNamespace,
			labelToNamespace,
		},
	)

	totalCrossRegionRequestCounts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNs,
			Subsystem: subSystemReverseProxy,
			Name:      "cross_region_requests_total",
			Help:      "Total count of cross region request",
		},
		[]string{
			labelFromNamespace,
			labelToNamespace,
		},
	)

	currentCrossRegionRequestCounts = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricNs,
			Subsystem: subSystemReverseProxy,
			Name:      "current_cross_region_request_counts",
			Help:      "Current count of cross region request",
		},
		[]string{
			labelFromNamespace,
			labelToNamespace,
		},
	)

	crossRegionRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricNs,
			Subsystem: subSystemReverseProxy,
			Name:      "cross_region_request_duration_seconds",
			Help:      "Total cross region request duration",
			Buckets:   prometheus.ExponentialBuckets(0.1, 2, 10),
		},
		[]string{
			labelFromNamespace,
			labelToNamespace,
		},
	)

	crossRegionRequestSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricNs,
			Subsystem: subSystemReverseProxy,
			Name:      "cross_region_request_size_bytes",
			Help:      "Cross region request size in bytes",
			Buckets:   prometheus.ExponentialBuckets(100, 2, 8),
		},
		[]string{
			labelFromNamespace,
			labelToNamespace,
		},
	)

	crossRegionResponseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricNs,
			Subsystem: subSystemReverseProxy,
			Name:      "cross_region_response_size_bytes",
			Help:      "Cross region response size in bytes",
			Buckets:   prometheus.ExponentialBuckets(100, 2, 8),
		},
		[]string{
			labelFromNamespace,
			labelToNamespace,
		},
	)
)

// external metrics duration only includes httpDo
func updateExternalMetrics(host string, start time.Time, resp *http.Response, err gopkg.CodeError) {
	if resp == nil && err == nil {
		return
	}

	HTTPCode := "0"
	eCode := "0"
	if err != nil {
		eCode = strconv.Itoa(err.ErrorCode())
	}
	if resp != nil {
		HTTPCode = getResponseMetricCode(resp)
	}

	externalUpstreamCounter.With(prometheus.Labels{
		labelHTTPCode:     HTTPCode,
		labelInternalCode: eCode,
		labelHost:         host,
	}).Inc()

	externalUpstreamDuration.With(prometheus.Labels{
		labelHTTPCode:     HTTPCode,
		labelInternalCode: eCode,
		labelHost:         host,
	}).Observe(time.Since(start).Seconds())
}

// internal metrics duration includes httpDo & m800DoPostProcessing
func updateInternalMetrics(host string, start time.Time, result *JsonResponse, err gopkg.CodeError) {
	if result == nil && err == nil {
		return
	}

	var code string
	if err != nil {
		code = strconv.Itoa(err.ErrorCode())
	} else {
		code = strconv.Itoa(result.Code)
	}

	internalUpstreamCounter.With(prometheus.Labels{
		labelInternalCode: code,
		labelHost:         host,
	}).Inc()
	internalUpstreamDuration.With(prometheus.Labels{
		labelInternalCode: code,
		labelHost:         host,
	}).Observe(time.Since(start).Seconds())
}
