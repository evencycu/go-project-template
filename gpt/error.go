package gpt

import (
	"net/http"

	"gitlab.com/cake/intercom"
)

const (
	CodeInternalServerError = 9990000

	CodeBadRequest = 9990001
	CodeForbidden  = 9990002

	CodeUpstreamSpecific = 9990100
)

func init() {
	// Status Bad Request Error 400
	_ = intercom.ErrorHttpStatusMapping.Set(CodeBadRequest, http.StatusBadRequest)

	// Status Forbidden Error 403
	_ = intercom.ErrorHttpStatusMapping.Set(CodeForbidden, http.StatusForbidden)

	// Status Internal Server Error 500
	_ = intercom.ErrorHttpStatusMapping.Set(CodeInternalServerError, http.StatusInternalServerError)
	_ = intercom.ErrorHttpStatusMapping.Set(CodeUpstreamSpecific, http.StatusInternalServerError)
}
