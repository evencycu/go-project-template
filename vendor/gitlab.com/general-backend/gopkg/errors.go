package gopkg

import (
	"fmt"
	"io"
	"strconv"

	"github.com/pkg/errors"
)

// CodeError - Code Error interface
type CodeError interface {
	ErrorCode() int
	Error() string
}

// ConstCodeError - 7 digits error code following a space and error message
type ConstCodeError string

// ErrorCode - return error code
func (cde ConstCodeError) ErrorCode() int {
	i, _ := strconv.Atoi(string(cde)[0:7])
	return i
}

// Error - return predefined error message
func (cde ConstCodeError) Error() string {
	return string(cde)[8:]
}

// Wrap - [Not Implemented] Wrap error message
func (cde ConstCodeError) Wrap(message string) error {
	return Wrap(cde, message)
}

// CarrierCodeError - 7 digits error code following a space and wrappable error message
type CarrierCodeError string

// ErrorCode - return error code
func (cde CarrierCodeError) ErrorCode() int {
	i, _ := strconv.Atoi(string(cde)[0:7])
	return i
}

// SetErrorCode - setup the error code and return new CDE
func (cde CarrierCodeError) SetErrorCode(code int) (CarrierCodeError, error) {
	if code < 1000000 || code > 9999999 {
		return cde, fmt.Errorf("error code should be 7 digits")
	}
	return CarrierCodeError(fmt.Sprintf("%d %s", code, cde.Error())), nil
}

// Error - return predefined error message
func (cde CarrierCodeError) Error() string {
	return string(cde)[8:]
}

// Wrap - append message to original one and return new CDE
func (cde CarrierCodeError) Wrap(message string) CarrierCodeError {
	return CarrierCodeError(fmt.Sprintf("%d %s", cde.ErrorCode(), cde.Error()+message))
}

// NewCarrierCodeError returns CarrierCodeError by given error code and message
func NewCarrierCodeError(code int, message string) CarrierCodeError {
	return CarrierCodeError(fmt.Sprintf("%d %s", code, message))
}

// NewCodeError returns CodeError by given error code and message (only accept 7-digits error code)
func NewCodeError(code int, message string) CodeError {
	return NewCarrierCodeError(code, message)
}

// AsCodeError - The Error "As" Code Error but more enhanced
type AsCodeError interface {
	CodeError
	CastOff() CodeError
	AsEqual(CodeError) bool
	Trace(string)
	Stack() error
	Format(s fmt.State, verb rune)
	Cause() error
}

// TraceCodeError -  Compatiable error stack trace & CodeError
// Contain a CodeError for "as" CodeError
// Using "error" & Wrap as a stack for tracing error history
type TraceCodeError struct {
	CodeError
	stack error
}

// NewTraceCodeError - Get a new *TraceCodeError from CodeError & external error
func NewTraceCodeError(der CodeError, err error) *TraceCodeError {
	return &TraceCodeError{
		CodeError: der,
		stack:     err,
	}
}

// NewTraceWithMsg  - One sugar function for NewTraceCodeError
func NewTraceWithMsg(der CodeError, msg string) *TraceCodeError {
	return &TraceCodeError{
		CodeError: der,
		stack:     NewError(msg),
	}
}

// Error - return error message
func (fde *TraceCodeError) Error() string {
	msg := fde.CodeError.Error()
	if fde.stack != nil {
		return fde.stack.Error() + ": " + msg
	}
	return msg
}

// CastOff - Unwrap TraceCodeError to return CodeError
func (fde *TraceCodeError) CastOff() CodeError {
	return fde.CodeError
}

// AsEqual - Return the CodeError is equal or not
func (fde *TraceCodeError) AsEqual(target CodeError) bool {
	return fde.CodeError == target
}

// Trace - Add message for trace
func (fde *TraceCodeError) Trace(message string) {
	if fde.stack == nil {
		fde.stack = NewError(message)
	} else {
		fde.stack = Wrap(fde.stack, message)
	}
}

// Stack - Return the stack
func (fde *TraceCodeError) Stack() error {
	return fde.stack
}

// Format - Implement the fmt.Formatter interface
// type Formatter interface {
//         Format(f State, c rune)
// }
// For fmt.Printf("%+v") to print stack trace
func (fde *TraceCodeError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v\n%+v\n", fde.stack, fde.CodeError)
			return
		}
		fallthrough
	case 's', 'q':
		io.WriteString(s, fde.Error())
	default:
		io.WriteString(s, fde.Error())
	}
}

// Cause - Implement Cause, Support to found the origin root Cause
func (fde *TraceCodeError) Cause() error {
	type causer interface {
		Cause() error
	}
	err := fde.stack
	for err != nil {
		cause, ok := err.(causer)
		if !ok {
			break
		}
		err = cause.Cause()
	}
	return err
}

func NewError(message string) error {
	return errors.New(message)
}

func Errorf(format string, args ...interface{}) error {
	return errors.Errorf(format, args...)
}

func Wrap(err error, message string) error {
	return errors.Wrap(err, message)
}

func Wrapf(err error, format string, args ...interface{}) error {
	return errors.Wrapf(err, format, args...)
}

func Cause(err error) error {
	return errors.Cause(err)
}
