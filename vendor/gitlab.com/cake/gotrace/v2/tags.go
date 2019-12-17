package gotrace

import (
	"net/http"
	"net/url"
)

type TagsMap struct {
	Method string
	URL    *url.URL
	Header http.Header
	Others map[string]string
}
