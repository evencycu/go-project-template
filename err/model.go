package err

import (
	"errors"
	"fmt"
)

var (
	// for sessionError
	ErrPermission = errors.New("permission denied")
	ErrClosed     = errors.New("connection already closed")

	sessionErrList = []error{ErrPermission, ErrClosed}

	// for numberError
	ErrTooLarge = errors.New("too large")
	ErrTooSmall = errors.New("too small")

	numberErrList = []error{ErrTooLarge, ErrTooSmall}
)

type sessionError struct {
	cause   string
	session string
	err     error
}

func (s sessionError) Error() string {
	return fmt.Sprintf("session %s down, cause by %s: %s", s.session, s.cause, s.err.Error())
}

func (s sessionError) Unwrap() error {
	return s.err
}

type numberError struct {
	number int
	err    error
}

func (n numberError) Error() string {
	return fmt.Sprintf("num %d: %s", n.number, n.err)
}

func (n numberError) Unwrap() error {
	return n.err
}
