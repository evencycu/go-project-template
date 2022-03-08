# gin-prometheus

Gin Web Framework Prometheus metrics exporter for M800

## Installation

`$ go get gitlab.com/cake/gin-prometheus`

## Usage

```go
package main

import (
	"github.com/gin-gonic/gin"
	ginprom "gitlab.com/cake/gin-prometheus"
)

func main() {
	r := gin.New()

	p := ginprom.NewPrometheus("gin")
	p.Use(r)

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, "Hello world!")
	})

	r.Run(":29090")
}
```

See the [example.go file](https://github.com/zsais/go-gin-prometheus/blob/master/example/example.go)

## Use Histogram type for request/response size and request duration

The default request size, response size and request duration metrics types
are `Summary`. In order to aggregate metrics with different labels, we provide
a option handler to use `Histogram` type. Note that it will cause high
prometheus server loading if the labels or bucket numbers are too many.
Learn more difference between `Summary` and `Histogram` here:

https://prometheus.io/docs/practices/histograms/

```go
package main

import (
	"github.com/gin-gonic/gin"
	ginprom "gitlab.com/cake/gin-prometheus"
)

func main() {
	r := gin.New()

	p, err := ginprom.NewPrometheus("gin",
		ginprom.HistogramMetrics("gin", ginprom.DefaultDurationBucket, ginprom.DefaultSizeBucket),
		ginprom.HistogramHandleFunc())
	if err != nil {
		return nil, err
	}

	p.Use(r)

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, "Hello world!")
	})

	r.Run(":29090")
}
```

## Prevent to use anonymous function as Handler

Since this framework use `gin.Context.HandlerName()` as the handler label value in metrics,
it is a better practice not to use anonymous function as `gin.HandlerFunc`. Following data
is a sample metrics when one runs `example/gin-metrics-custom/custom.go` server. Note that
when one request to `GET /anonymous` path, it will record handler name as `main.main.func1`.
First `main` is the package name, second `main` is the function name and `func1` means the
first anonymous function in this main function.

Also, we suggest to implement `func (engine *Engine) NoRoute(handlers ...HandlerFunc)` and
`func (engine *Engine) NoMethod(handlers ...HandlerFunc)` to prevent generating misleading
handler label value in metrics.

```
# HELP gin_request_duration_seconds The HTTP request latencies in seconds.
# TYPE gin_request_duration_seconds histogram
gin_request_duration_seconds_bucket{code="2xx",handler="main.main.func1",method="GET",le="0.1"} 1
gin_request_duration_seconds_bucket{code="2xx",handler="main.main.func1",method="GET",le="0.2"} 1
gin_request_duration_seconds_bucket{code="2xx",handler="main.main.func1",method="GET",le="0.4"} 1
gin_request_duration_seconds_bucket{code="2xx",handler="main.main.func1",method="GET",le="0.8"} 1
gin_request_duration_seconds_bucket{code="2xx",handler="main.main.func1",method="GET",le="1.6"} 1
gin_request_duration_seconds_bucket{code="2xx",handler="main.main.func1",method="GET",le="3.2"} 1
gin_request_duration_seconds_bucket{code="2xx",handler="main.main.func1",method="GET",le="6.4"} 1
gin_request_duration_seconds_bucket{code="2xx",handler="main.main.func1",method="GET",le="+Inf"} 1
gin_request_duration_seconds_sum{code="2xx",handler="main.main.func1",method="GET"} 3.1355e-05
gin_request_duration_seconds_count{code="2xx",handler="main.main.func1",method="GET"} 1
gin_request_duration_seconds_bucket{code="2xx",handler="main.namedHandler",method="GET",le="0.1"} 1
gin_request_duration_seconds_bucket{code="2xx",handler="main.namedHandler",method="GET",le="0.2"} 1
gin_request_duration_seconds_bucket{code="2xx",handler="main.namedHandler",method="GET",le="0.4"} 1
gin_request_duration_seconds_bucket{code="2xx",handler="main.namedHandler",method="GET",le="0.8"} 1
gin_request_duration_seconds_bucket{code="2xx",handler="main.namedHandler",method="GET",le="1.6"} 1
gin_request_duration_seconds_bucket{code="2xx",handler="main.namedHandler",method="GET",le="3.2"} 1
gin_request_duration_seconds_bucket{code="2xx",handler="main.namedHandler",method="GET",le="6.4"} 1
gin_request_duration_seconds_bucket{code="2xx",handler="main.namedHandler",method="GET",le="+Inf"} 1
gin_request_duration_seconds_sum{code="2xx",handler="main.namedHandler",method="GET"} 5.5448e-05
gin_request_duration_seconds_count{code="2xx",handler="main.namedHandler",method="GET"} 1
gin_request_duration_seconds_bucket{code="4xx",handler="main.noRoute",method="GET",le="0.1"} 2
gin_request_duration_seconds_bucket{code="4xx",handler="main.noRoute",method="GET",le="0.2"} 2
gin_request_duration_seconds_bucket{code="4xx",handler="main.noRoute",method="GET",le="0.4"} 2
gin_request_duration_seconds_bucket{code="4xx",handler="main.noRoute",method="GET",le="0.8"} 2
gin_request_duration_seconds_bucket{code="4xx",handler="main.noRoute",method="GET",le="1.6"} 2
gin_request_duration_seconds_bucket{code="4xx",handler="main.noRoute",method="GET",le="3.2"} 2
gin_request_duration_seconds_bucket{code="4xx",handler="main.noRoute",method="GET",le="6.4"} 2
gin_request_duration_seconds_bucket{code="4xx",handler="main.noRoute",method="GET",le="+Inf"} 2
gin_request_duration_seconds_sum{code="4xx",handler="main.noRoute",method="GET"} 6.6965e-05
gin_request_duration_seconds_count{code="4xx",handler="main.noRoute",method="GET"} 2
```
