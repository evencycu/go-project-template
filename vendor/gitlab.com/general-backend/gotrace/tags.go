package gotrace

import (
	opentracing "github.com/opentracing/opentracing-go"
	ext "github.com/opentracing/opentracing-go/ext"

	"net/http"
	"net/url"
)

type TagsMap struct {
	Method string
	URL    *url.URL
	Header http.Header
	Others map[string]string
}

func SetRPCClientTag(span opentracing.Span) {
	ext.SpanKindRPCClient.Set(span)
}
