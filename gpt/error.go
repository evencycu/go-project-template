package gpt

import (
	"net/http"

	"gitlab.com/cake/intercom"
)

const (
	CodeInternalServerError = 9990000

	CodeBadRequest = 9990001
)

func init() {
	// Status Internal Server Error 400
	_ = intercom.ErrorHttpStatusMapping.Set(CodeBadRequest, http.StatusBadRequest)

	// Status Internal Server Error 500
	_ = intercom.ErrorHttpStatusMapping.Set(CodeInternalServerError, http.StatusInternalServerError)
}
