package se

import (
	"net/http"

	"gitlab.com/cake/gopkg"
	"gitlab.com/cake/intercom"
)

const (
	// General Error
	CodeInternalServerError = 7210000
	CodeBadGateway          = 7210001
	CodeRouteNotFound       = 7210002

	// Request Error
	CodeBadRequest                   = 7210101
	CodeAnonymous                    = 7210102
	CodeSenderNumberInvalidFormat    = 7210103
	CodeRecipientNumberInvalidFormat = 7210104

	// Dependency Error
	CodeSendSMSFail = 7210201
	CodeRedisError  = 7210202

	// Business Logic Error
	CodeSMSDisabled         = 7210301
	CodeSMSRateLimitReached = 7210302
)

var (
	ErrSMSDisabled         = gopkg.NewCodeError(CodeSMSDisabled, "SMS disabled")
	ErrSMSRateLimitReached = gopkg.NewCodeError(CodeSMSRateLimitReached, "SMS rate limit reached")
	ErrAnms                = gopkg.NewCodeError(CodeAnonymous, "anonymous user not allowed")

	ErrRedisError = gopkg.NewCodeError(CodeRedisError, "database failure")
)

func init() {
	// Status Bad Request 400
	_ = intercom.ErrorHttpStatusMapping.Set(CodeBadRequest, http.StatusBadRequest)
	_ = intercom.ErrorHttpStatusMapping.Set(CodeSenderNumberInvalidFormat, http.StatusBadRequest)
	_ = intercom.ErrorHttpStatusMapping.Set(CodeRecipientNumberInvalidFormat, http.StatusBadRequest)

	// Status Unauthorized 401
	_ = intercom.ErrorHttpStatusMapping.Set(CodeAnonymous, http.StatusUnauthorized)

	// Status Forbidden 403
	_ = intercom.ErrorHttpStatusMapping.Set(CodeSMSDisabled, http.StatusForbidden)
	_ = intercom.ErrorHttpStatusMapping.Set(CodeSMSRateLimitReached, http.StatusForbidden)

	// Status Internal Server Error 500
	_ = intercom.ErrorHttpStatusMapping.Set(CodeRedisError, http.StatusInternalServerError)
}
