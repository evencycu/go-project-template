package goctx

import (
	"github.com/opentracing/opentracing-go"
)

func (c *MapContext) GetSpan() opentracing.Span {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return opentracing.SpanFromContext(c.Context)
}

func (c *MapContext) StartSpanFromContext(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	c.mu.Lock()
	defer c.mu.Unlock()
	sp, gCtx := opentracing.StartSpanFromContext(c.Context, operationName, opts...)
	c.Context = gCtx
	return sp
}

func (c *MapContext) SetSpan(span opentracing.Span) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Context = opentracing.ContextWithSpan(c.Context, span)
}
