package gotrace

import (
	"net"
	"net/http"
	"strconv"

	opentracing "github.com/opentracing/opentracing-go"
	ext "github.com/opentracing/opentracing-go/ext"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	// blank import for better readibility
	_ "github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	metrics "github.com/uber/jaeger-lib/metrics"
	"gitlab.com/general-backend/goctx"
)

type SpanReference string

func InitJaeger(componentName string, samplerConf *jaegercfg.SamplerConfig, reporterConf *jaegercfg.ReporterConfig) error {
	cfg := jaegercfg.Configuration{
		ServiceName: componentName,
		Sampler:     samplerConf,
		Reporter:    reporterConf,
	}
	// TODO: Add logger
	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory
	tracer, _, err := cfg.NewTracer(
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
		jaegercfg.ZipkinSharedRPCSpan(true),
	)
	if err != nil {
		return err
	}
	opentracing.SetGlobalTracer(tracer)
	return nil
}

func InitZipkin(collector zipkin.Collector, sampler zipkin.Sampler, host, componentName string) error {
	tracer, err := zipkin.NewTracer(
		zipkin.NewRecorder(collector, false, host, componentName),
		zipkin.ClientServerSameSpan(true),
		zipkin.TraceID128Bit(true),
		zipkin.WithSampler(sampler),
	)
	if err != nil {
		return err
	}
	opentracing.SetGlobalTracer(tracer)
	return nil
}

// ExtractSpanFromContext create span from context value
func ExtractSpanFromContext(spanName string, ctx goctx.Context) opentracing.Span {
	var sp opentracing.Span

	wireContext, err := opentracing.GlobalTracer().Extract(
		opentracing.TextMap,
		opentracing.TextMapCarrier(ctx.LogKeyMap()))
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

	return sp
}

func ExtractSpanFromTagsMap(spanName string, tags *TagsMap) opentracing.Span {
	var sp opentracing.Span
	wireContext, err := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(tags.Header))
	if err != nil {
		sp = opentracing.StartSpan(spanName)
	} else {
		sp = opentracing.StartSpan(
			spanName,
			ext.RPCServerOption(wireContext))
	}
	attachSpanTags(sp, tags)
	return sp
}

func ExtractSpanFromReq(spanName string, req *http.Request) opentracing.Span {
	var sp opentracing.Span
	wireContext, err := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header))
	if err != nil {
		sp = opentracing.StartSpan(spanName)
	} else {
		sp = opentracing.StartSpan(
			spanName,
			ext.RPCServerOption(wireContext))
	}
	tags := &TagsMap{
		Method: req.Method,
		URL:    req.URL,
		Header: req.Header,
	}
	attachSpanTags(sp, tags)
	return sp
}

func CreateSpanFromReq(spaneName string, parentSpan opentracing.Span, relation SpanReference, req *http.Request) opentracing.Span {
	tags := &TagsMap{
		Method: req.Method,
		URL:    req.URL,
		Header: req.Header,
	}
	return CreateSpan(spaneName, parentSpan, relation, tags)
}

func CreateSpanByContext(spanName string, ctx goctx.Context, relation SpanReference, tags *TagsMap) opentracing.Span {
	if ctx == nil {
		return CreateSpan(spanName, nil, ReferenceRoot, tags)
	}
	parentSpanTmp := ctx.Get(goctx.LogKeyTrace)
	if parentSpanTmp == nil {
		return CreateSpan(spanName, nil, ReferenceRoot, tags)
	}
	parentSpan, ok := parentSpanTmp.(opentracing.Span)
	if !ok {
		return CreateSpan(spanName, nil, ReferenceRoot, tags)
	}
	return CreateSpan(spanName, parentSpan, relation, tags)
}

func CreateSpan(spanName string, parentSpan opentracing.Span, relation SpanReference, tags *TagsMap) opentracing.Span {
	var sp opentracing.Span

	switch relation {
	case ReferenceRoot:
		sp = opentracing.StartSpan(spanName)
		attachSpanTags(sp, tags)
	case ReferenceChildOf:
		sp = opentracing.StartSpan(
			spanName,
			opentracing.ChildOf(parentSpan.Context()))
		attachSpanTags(sp, tags)
	case ReferenceFollowsFrom:
		sp = opentracing.StartSpan(
			spanName,
			opentracing.FollowsFrom(parentSpan.Context()))
		attachSpanTags(sp, tags)
	}
	return sp
}

func attachSpanTags(sp opentracing.Span, tags *TagsMap) {
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
	err := sp.Tracer().Inject(sp.Context(),
		opentracing.TextMap,
		opentracing.HTTPHeadersCarrier(header))
	return err
}
