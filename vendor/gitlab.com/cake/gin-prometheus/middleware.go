package ginmetrics

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	MetricCollector prometheus.Collector
	ID              string
	Name            string
	Description     string
	Type            string
	Args            []string
	Opts            interface{}
}

type GinPrometheus struct {
	ginHandleFunc gin.HandlerFunc

	MetricsList map[string]*Metrics

	Cnt          map[string]prometheus.Counter
	CntVec       map[string]*prometheus.CounterVec
	Gauge        map[string]prometheus.Gauge
	GaugeVec     map[string]*prometheus.GaugeVec
	Histogram    map[string]prometheus.Histogram
	HistogramVec map[string]*prometheus.HistogramVec
	Summary      map[string]prometheus.Summary
	SummaryVec   map[string]*prometheus.SummaryVec

	router        *gin.Engine
	listenAddress string
	MetricsPath   string
}

func (p *GinPrometheus) SetHandlerFunc(h gin.HandlerFunc) {
	p.ginHandleFunc = h
}

func NewPrometheus(subsystem string, options ...func(*GinPrometheus)) (p *GinPrometheus, err error) {

	p = &GinPrometheus{}

	p.addDefaultMetrics(subsystem)
	p.SetHandlerFunc(p.DefaultHandleFunc())
	p.MetricsPath = defaultMetricPath

	for _, option := range options {
		option(p)
	}

	if err = p.registerMetrics(); err != nil {
		return nil, err
	}

	return
}

func (p *GinPrometheus) addDefaultMetrics(subsystem string) {
	p.MetricsList = make(map[string]*Metrics)

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
		Opts: prometheus.SummaryOpts{
			Subsystem: subsystem,
			Name:      "request_duration_seconds",
			Help:      "The HTTP request latencies in seconds.",
		},
		Type: TypeSummaryVec,
		Args: []string{"code", "method", "handler"},
	}

	p.MetricsList[KeyReqSz] = &Metrics{
		Opts: prometheus.SummaryOpts{
			Subsystem: subsystem,
			Name:      "request_size_bytes",
			Help:      "The HTTP request sizes in bytes.",
		},
		Type: TypeSummaryVec,
		Args: []string{"code", "method", "handler"},
	}

	p.MetricsList[KeyResSz] = &Metrics{
		Opts: prometheus.SummaryOpts{
			Subsystem: subsystem,
			Name:      "response_size_bytes",
			Help:      "The HTTP response sizes in bytes.",
		},
		Type: TypeSummaryVec,
		Args: []string{"code", "method", "handler"},
	}
}

func (p *GinPrometheus) registerMetrics() (err error) {

	p.Cnt = make(map[string]prometheus.Counter)
	p.CntVec = make(map[string]*prometheus.CounterVec)
	p.Gauge = make(map[string]prometheus.Gauge)
	p.GaugeVec = make(map[string]*prometheus.GaugeVec)
	p.Histogram = make(map[string]prometheus.Histogram)
	p.HistogramVec = make(map[string]*prometheus.HistogramVec)
	p.Summary = make(map[string]prometheus.Summary)
	p.SummaryVec = make(map[string]*prometheus.SummaryVec)

	for key, metric := range p.MetricsList {

		switch metric.Type {
		case TypeCounter:
			opts, ok := metric.Opts.(prometheus.CounterOpts)
			if !ok {
				return errors.New(fmt.Sprintf(strAssert, key, metric.Type))
			}
			p.Cnt[key] = prometheus.NewCounter(opts)
			err = prometheus.Register(p.Cnt[key])
			if err != nil {
				return
			}

		case TypeCounterVec:
			opts, ok := metric.Opts.(prometheus.CounterOpts)
			if !ok {
				return errors.New(fmt.Sprintf(strAssert, key, metric.Type))
			}
			p.CntVec[key] = prometheus.NewCounterVec(opts, metric.Args)
			err = prometheus.Register(p.CntVec[key])
			if err != nil {
				return
			}

		case TypeGauge:
			opts, ok := metric.Opts.(prometheus.GaugeOpts)
			if !ok {
				return errors.New(fmt.Sprintf(strAssert, key, metric.Type))
			}
			p.Gauge[key] = prometheus.NewGauge(opts)
			err = prometheus.Register(p.Gauge[key])
			if err != nil {
				return
			}

		case TypeGaugeVec:
			opts, ok := metric.Opts.(prometheus.GaugeOpts)
			if !ok {
				return errors.New(fmt.Sprintf(strAssert, key, metric.Type))
			}
			p.GaugeVec[key] = prometheus.NewGaugeVec(opts, metric.Args)
			err = prometheus.Register(p.GaugeVec[key])
			if err != nil {
				return
			}

		case TypeHistogram:
			opts, ok := metric.Opts.(prometheus.HistogramOpts)
			if !ok {
				return errors.New(fmt.Sprintf(strAssert, key, metric.Type))
			}
			p.Histogram[key] = prometheus.NewHistogram(opts)
			err = prometheus.Register(p.Histogram[key])
			if err != nil {
				return
			}

		case TypeHistogramVec:
			opts, ok := metric.Opts.(prometheus.HistogramOpts)
			if !ok {
				return errors.New(fmt.Sprintf(strAssert, key, metric.Type))
			}
			p.HistogramVec[key] = prometheus.NewHistogramVec(opts, metric.Args)
			err = prometheus.Register(p.HistogramVec[key])
			if err != nil {
				return
			}

		case TypeSummary:
			opts, ok := metric.Opts.(prometheus.SummaryOpts)
			if !ok {
				return errors.New(fmt.Sprintf(strAssert, key, metric.Type))
			}
			p.Summary[key] = prometheus.NewSummary(opts)
			err = prometheus.Register(p.Summary[key])
			if err != nil {
				return
			}

		case TypeSummaryVec:
			opts, ok := metric.Opts.(prometheus.SummaryOpts)
			if !ok {
				return errors.New(fmt.Sprintf(strAssert, key, metric.Type))
			}
			p.SummaryVec[key] = prometheus.NewSummaryVec(opts, metric.Args)
			err = prometheus.Register(p.SummaryVec[key])
			if err != nil {
				return
			}
		}

	}
	return
}

// Use adds the middleware to a gin engine.
func (p *GinPrometheus) Use(e *gin.Engine) {
	e.Use(p.ginHandleFunc)
	e.GET(p.MetricsPath, prometheusHandler())
}

// SetMetricsPath set metrics paths
func (p *GinPrometheus) SetMetricsPath(e *gin.Engine) {

	if p.listenAddress != "" {
		p.router.GET(p.MetricsPath, prometheusHandler())
		p.runServer()
	} else {
		e.GET(p.MetricsPath, prometheusHandler())
	}
}

func (p *GinPrometheus) runServer() {
	if p.listenAddress != "" {
		go p.router.Run(p.listenAddress)
	}
}

func prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func (p *GinPrometheus) DefaultHandleFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.String() == p.MetricsPath {
			c.Next()
			return
		}

		var status string

		timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
			p.SummaryVec[KeyReqDur].WithLabelValues(status, c.Request.Method, c.HandlerName()).Observe(v)
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
		resSz := float64(c.Writer.Size())

		p.CntVec[KeyReqCnt].WithLabelValues(status, c.Request.Method, c.HandlerName()).Inc()
		p.SummaryVec[KeyReqSz].WithLabelValues(status, c.Request.Method, c.HandlerName()).Observe(float64(reqSz))
		p.SummaryVec[KeyResSz].WithLabelValues(status, c.Request.Method, c.HandlerName()).Observe(resSz)
	}
}

// From https://github.com/DanielHeckrath/gin-prometheus/blob/master/gin_prometheus.go
func ComputeApproximateRequestSize(r *http.Request) int {
	s := 0
	if r.URL != nil {
		s = len(r.URL.String())
	}

	s += len(r.Method)
	s += len(r.Proto)
	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}
	s += len(r.Host)

	// N.B. r.Form and r.MultipartForm are assumed to be included in r.URL.

	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}
	return s
}
