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

// CreateSpan extract span from given context first, if so, return it
// else create a new span (from empty, or by context header) and inject to context
func CreateSpan(ctx goctx.Context, spanName string) (sp opentracing.Span, needFinish bool) {
	sp = ctx.GetSpan()
	if sp != nil {
		return
	}
	needFinish = true
	wireContext, err := opentracing.GlobalTracer().Extract(
		opentracing.TextMap,
		opentracing.TextMapCarrier(ctx.HeaderKeyMap()))
	if err != nil {
		sp = ctx.StartSpanFromContext(spanName)
	} else {
		sp = ctx.StartSpanFromContext(spanName, opentracing.ChildOf(wireContext))
	}
	for k, v := range ctx.Map() {
		sp.SetTag(k, v)
	}
	return
}

func CreateChildOfSpan(ctx goctx.Context, spanName string) (sp opentracing.Span) {
	sp = ctx.StartSpanFromContext(spanName)
	return
}

func CreateFollowsFromSpan(ctx goctx.Context, spanName string) (sp opentracing.Span) {
	parentSpan := ctx.GetSpan()
	if parentSpan != nil {
		sp = opentracing.StartSpan(spanName, opentracing.FollowsFrom(parentSpan.Context()))
		ctx.SetSpan(sp)
		return
	}
	wireContext, err := opentracing.GlobalTracer().Extract(
		opentracing.TextMap,
		opentracing.TextMapCarrier(ctx.HeaderKeyMap()))
	if err != nil {
		sp = opentracing.StartSpan(spanName)
	} else {
		sp = opentracing.StartSpan(spanName, opentracing.FollowsFrom(wireContext))
	}
	ctx.SetSpan(sp)
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
