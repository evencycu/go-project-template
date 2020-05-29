package gopkg

import (
	"fmt"
	"strconv"
)

// CodeError - Code Error interface
type CodeError interface {
	ErrorCode() int
	ErrorMsg() string
	Error() string
	FullError() string
}

// ConstCodeError - 7 digits error code following a space and error message
type ConstCodeError string

// ErrorCode - return error code
func (cde ConstCodeError) ErrorCode() int {
	i, _ := strconv.Atoi(string(cde)[0:7])
	return i
}

// ErrorMsg - return predefined error message
func (cde ConstCodeError) ErrorMsg() string {
	return string(cde)[8:]
}

// Error - return formatted error code and error message
func (cde ConstCodeError) Error() string {
	return string(cde)
}

// FullError - return formatted error code and error message
func (cde ConstCodeError) FullError() string {
	return string(cde)
}

// CarrierCodeError - 7 digits error code following a space and wrapable error message
// If the ErrMsg is empty, it represents this error happens from upstream
type CarrierCodeError struct {
	ErrCode int
	ErrMsg  string
	WrapErr error
}

// ErrorCode - return error code
func (cde CarrierCodeError) ErrorCode() int {
	if cde.ErrCode == 0 {
		if we, ok := cde.WrapErr.(CodeError); ok {
			return we.ErrorCode()
		}
	}
	return cde.ErrCode
}

// SetErrorCode - setup the error code and return new CarrierCodeError instance
func (cde CarrierCodeError) SetErrorCode(code int) (CarrierCodeError, error) {
	if code < 1000000 || code > 9999999 {
		return cde, fmt.Errorf("error code should be 7 digits")
	}
	cde.ErrCode = code
	return cde, nil
}

// ErrorMsg - return error message
func (cde CarrierCodeError) ErrorMsg() string {
	if cde.ErrMsg == "" {
		if we, ok := cde.WrapErr.(CarrierCodeError); ok {
			return we.ErrorMsg()
		}
	}
	return cde.ErrMsg
}

// Error - return formatted error code and error message
func (cde CarrierCodeError) Error() string {
	emptyString := ""
	if cde.ErrMsg == emptyString {
		if cde.WrapErr != nil {
			return cde.WrapErr.Error()
		}
		return emptyString
	}
	if cde.WrapErr != nil {
		return fmt.Sprintf("%07d %s: %v", cde.ErrCode, cde.ErrMsg, cde.WrapErr)
	}
	return fmt.Sprintf("%07d %s", cde.ErrCode, cde.ErrMsg)
}

// FullError - return formatted error code and error message
func (cde CarrierCodeError) FullError() string {
	return cde.Error()
}

// WrappedError - return wrappedError
func (cde CarrierCodeError) WrappedError() string {
	if cde.WrapErr == nil {
		return ""
	}
	return cde.WrapErr.Error()
}

// As implements golang 1.13 errors interface
// It retains extendibility to wrap more type of error
func (cde CarrierCodeError) As(err interface{}) bool {
	switch x := err.(type) {
	case *CodeError:
		*x = cde
	case *CarrierCodeError:
		*x = cde
	default:
		return false
	}
	return true
}

// Unwrap implements golang 1.13 error interface
func (cde CarrierCodeError) Unwrap() error {
	return cde.WrapErr
}

// NewWrappedCarrierCodeError returns CarrierCodeError by given error code, message and error
func NewWrappedCarrierCodeError(code int, message string, err error) CarrierCodeError {
	if ce, ok := err.(CarrierCodeError); ok {
		if ce.ErrMsg == "" {
			ce.ErrCode = code
			ce.ErrMsg = message
			return ce
		}
	}
	return CarrierCodeError{
		ErrCode: code,
		ErrMsg:  message,
		WrapErr: err,
	}
}

// NewCarrierCodeError returns CarrierCodeError by given error code and message
func NewCarrierCodeError(code int, message string) CarrierCodeError {
	return CarrierCodeError{
		ErrCode: code,
		ErrMsg:  message,
	}
}

// NewWrappedCodeError returns CodeError by given error code, message and error (only accept 7-digits error code)
func NewWrappedCodeError(code int, message string, wrappedErr error) CodeError {
	return NewWrappedCarrierCodeError(code, message, wrappedErr)
}

// NewCodeError returns CodeError by given error code and message (only accept 7-digits error code)
func NewCodeError(code int, message string) CodeError {
	return NewCarrierCodeError(code, message)
}
