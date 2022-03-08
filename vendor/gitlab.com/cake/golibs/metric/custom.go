package metric

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"
)

func HistogramMetrics(subsystem string, durationBucket, sizeBucket []float64) func(*GinPrometheus) {
	return func(p *GinPrometheus) {

		p.MetricsList[KeyReqCnt] = &Metrics{
			Opts: prometheus.CounterOpts{
				Subsystem: subsystem,
				Name:      "requests_total",
				Help:      "How many HTTP requests processed, partitioned by status code and HTTP method.",
			},
			Type: TypeCounterVec,
			Args: []string{"code", "method", "handler"},
		}

		p.MetricsList[KeyReqDur] = &Metrics{
			Opts: prometheus.HistogramOpts{
				Subsystem: subsystem,
				Name:      "request_duration_seconds",
				Help:      "The HTTP request latencies in seconds.",
				Buckets:   durationBucket,
			},
			Type: TypeHistogramVec,
			Args: []string{"code", "method", "handler"},
		}

		p.MetricsList[KeyReqSz] = &Metrics{
			Opts: prometheus.HistogramOpts{
				Subsystem: subsystem,
				Name:      "request_size_bytes",
				Help:      "The HTTP request sizes in bytes.",
				Buckets:   sizeBucket,
			},
			Type: TypeHistogramVec,
			Args: []string{"code", "method", "handler"},
		}

		p.MetricsList[KeyResSz] = &Metrics{
			Opts: prometheus.HistogramOpts{
				Subsystem: subsystem,
				Name:      "response_size_bytes",
				Help:      "The HTTP response sizes in bytes.",
				Buckets:   sizeBucket,
			},
			Type: TypeHistogramVec,
			Args: []string{"code", "method", "handler"},
		}

		p.MetricsList[KeyReqDurAll] = &Metrics{
			Opts: prometheus.SummaryOpts{
				Subsystem:  subsystem,
				Name:       "request_overall_duration_seconds",
				Help:       "The HTTP overall request latencies in seconds.",
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
			},
			Type: TypeSummary,
		}
	}
}

func HistogramHandleFunc() func(p *GinPrometheus) {
	return func(p *GinPrometheus) {
		p.SetHandlerFunc(func(c *gin.Context) {
			if c.Request.URL.String() == p.MetricsPath {
				c.Next()
				return
			}

			var (
				status      string
				handlerName string

				traceID    = trace.SpanContextFromContext(c.Request.Context()).TraceID()
				traceLabel = prometheus.Labels{KeyTraceID: traceID.String()}
			)

			timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
				if observer, ok := p.HistogramVec[KeyReqDur].WithLabelValues(status, c.Request.Method, handlerName).(prometheus.ExemplarObserver); ok && traceID.IsValid() {
					observer.ObserveWithExemplar(v, traceLabel)
				} else {
					p.HistogramVec[KeyReqDur].WithLabelValues(status, c.Request.Method, handlerName).Observe(v)
				}
				p.Summary[KeyReqDurAll].Observe(v)
			}))
			defer timer.ObserveDuration()

			reqSz := ComputeApproximateRequestSize(c.Request)

			c.Next()

			switch {
			case c.Writer.Status() >= 500:
				status = "5xx"
			case c.Writer.Status() >= 400: // Client error.
				status = "4xx"
			case c.Writer.Status() >= 300: // Redirection.
				status = "3xx"
			case c.Writer.Status() >= 200: // Success.
				status = "2xx"
			default: // Informational.
				status = strconv.Itoa(c.Writer.Status())
			}
			handlerName = c.HandlerName()
			resSz := float64(c.Writer.Size())

			if adder, ok := p.CntVec[KeyReqCnt].WithLabelValues(status, c.Request.Method, handlerName).(prometheus.ExemplarAdder); ok && traceID.IsValid() {
				adder.AddWithExemplar(1, traceLabel)
			} else {
				p.CntVec[KeyReqCnt].WithLabelValues(status, c.Request.Method, handlerName).Inc()
			}
			p.HistogramVec[KeyReqSz].WithLabelValues(status, c.Request.Method, handlerName).Observe(float64(reqSz))
			p.HistogramVec[KeyResSz].WithLabelValues(status, c.Request.Method, handlerName).Observe(resSz)
		})
	}
}
