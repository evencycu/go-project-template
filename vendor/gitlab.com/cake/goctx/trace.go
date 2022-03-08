package goctx

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const (
	tracerName    = "gitlab.com/cake/goctx"
	tracerVersion = "unset"
)

var tracer = otel.GetTracerProvider().Tracer(
	tracerName,
	oteltrace.WithInstrumentationVersion(tracerVersion),
)

func (c *MapContext) GetSpan() oteltrace.Span {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return oteltrace.SpanFromContext(c.Context)
}

func (c *MapContext) GetBaggage() baggage.Baggage {
	c.mu.Lock()
	defer c.mu.Unlock()
	return baggage.FromContext(c.Context)
}

func (c *MapContext) StartSpanFromContext(spanName string, opts ...oteltrace.SpanStartOption) oteltrace.Span {
	c.mu.Lock()
	defer c.mu.Unlock()
	newCtx, sp := tracer.Start(c.Context, spanName, opts...)
	c.Context = newCtx
	return sp
}

func (c *MapContext) SetSpan(span oteltrace.Span) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Context = oteltrace.ContextWithSpan(c.Context, span)
}

func (c *MapContext) SetBaggage(b baggage.Baggage) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Context = baggage.ContextWithBaggage(c.Context, b)
}
