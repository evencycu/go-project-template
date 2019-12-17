package gotrace

import (
	"io"
	"net"
	"net/http"
	"strconv"

	opentracing "github.com/opentracing/opentracing-go"
	ext "github.com/opentracing/opentracing-go/ext"
	// blank import for better readibility
	_ "github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	metrics "github.com/uber/jaeger-lib/metrics"
	"gitlab.com/cake/goctx"
)

func InitJaeger(componentName string, samplerConf *jaegercfg.SamplerConfig, reporterConf *jaegercfg.ReporterConfig) (io.Closer, error) {
	cfg := jaegercfg.Configuration{
		ServiceName: componentName,
		Sampler:     samplerConf,
		Reporter:    reporterConf,
	}
	// TODO: Add logger
	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory
	tracer, closer, err := cfg.NewTracer(
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
		jaegercfg.ZipkinSharedRPCSpan(true),
	)
	if err != nil {
		return nil, err
	}
	opentracing.SetGlobalTracer(tracer)
	return closer, nil
}

// // CreateSpanFromReq extract span from given http request and inject to context
// // always need finish
// func CreateSpanFromReq(ctx goctx.Context, spanName string, req *http.Request) opentracing.Span {
// 	var sp opentracing.Span
// 	wireContext, err := opentracing.GlobalTracer().Extract(
// 		opentracing.HTTPHeaders,
// 		opentracing.HTTPHeadersCarrier(req.Header))
// 	if err != nil {
// 		sp = opentracing.StartSpan(spanName)
// 	} else {
// 		sp = opentracing.StartSpan(
// 			spanName,
// 			ext.RPCServerOption(wireContext))
// 	}
// 	tags := &TagsMap{
// 		Method: req.Method,
// 		URL:    req.URL,
// 		Header: req.Header,
// 	}
// 	AttachHttpTags(sp, tags)
// 	InjectContext(sp, ctx)
// 	return sp
// }

// CreateSpan extract span from given context first, if so, return it
// else create a new span (from empty, or by context header) and inject to context
func CreateSpan(ctx goctx.Context, spanName string) (sp opentracing.Span, needFinish bool) {
	var ok bool
	sp, ok = ctx.Get(goctx.LogKeyTrace).(opentracing.Span)
	if ok {
		return
	}
	needFinish = true
	wireContext, err := opentracing.GlobalTracer().Extract(
		opentracing.TextMap,
		opentracing.TextMapCarrier(ctx.HeaderKeyMap()))
	if err != nil {
		sp = opentracing.StartSpan(spanName)
	} else {
		sp = opentracing.StartSpan(
			spanName,
			ext.RPCServerOption(wireContext))
	}
	for k, v := range ctx.Map() {
		sp.SetTag(k, v)
	}
	InjectContext(sp, ctx)
	return
}

func CreateChildOfSpan(ctx goctx.Context, spanName string) (sp opentracing.Span) {
	parentSpan, ok := ctx.Get(goctx.LogKeyTrace).(opentracing.Span)
	if ok {
		sp = opentracing.StartSpan(
			spanName,
			opentracing.ChildOf(parentSpan.Context()))
		_ = InjectContext(sp, ctx)
		return sp
	}
	sp, _ = CreateSpan(ctx, spanName)
	return
}

func CreateFollowsFromSpan(ctx goctx.Context, spanName string) (sp opentracing.Span) {
	parentSpan, ok := ctx.Get(goctx.LogKeyTrace).(opentracing.Span)
	if ok {
		sp = opentracing.StartSpan(
			spanName,
			opentracing.FollowsFrom(parentSpan.Context()))
		_ = InjectContext(sp, ctx)
		return
	}
	sp, _ = CreateSpan(ctx, spanName)
	return
}

func AttachHttpTags(sp opentracing.Span, tags *TagsMap) {
	if tags == nil {
		return
	}
	if len(tags.Method) != 0 {
		ext.HTTPMethod.Set(sp, tags.Method)
	}
	if tags.URL != nil {
		ext.HTTPUrl.Set(sp, tags.URL.String())
		if host, portString, err := net.SplitHostPort(tags.URL.Host); err == nil {
			ext.PeerHostname.Set(sp, host)
			if port, err := strconv.Atoi(portString); err != nil {
				ext.PeerPort.Set(sp, uint16(port))
			}
		} else {
			ext.PeerHostname.Set(sp, tags.URL.Host)
		}
	}

	if tags.Others != nil {
		for k, v := range tags.Others {
			sp.SetTag(k, v)
		}
	}
}

func InjectSpan(sp opentracing.Span, header http.Header) error {
	ext.SpanKindRPCClient.Set(sp)
	err := sp.Tracer().Inject(sp.Context(),
		opentracing.TextMap,
		opentracing.HTTPHeadersCarrier(header))
	return err
}

func InjectContext(sp opentracing.Span, ctx goctx.Context) error {
	// no use inject because ctx is not opentracing.TextMap
	// ctx is Set(string,interface{})
	// opentracing.TextMap is Set(string,string)
	ctx.Set(goctx.LogKeyTrace, sp)
	return nil
}
