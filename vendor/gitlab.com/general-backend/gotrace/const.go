package gotrace

const (
	TagError      = "error"
	TagHTTPStatus = "http.status_code"
)

const (
	ReferenceRoot        = SpanReference("root")
	ReferenceChildOf     = SpanReference("childOf")
	ReferenceFollowsFrom = SpanReference("followsFrom")
)
