package errors

import "github.com/pkg/errors"

// ErrorWithStack implements errors.StackTrace to be compatible with sentry
type ErrorWithStack struct {
	Stack errors.StackTrace
	Err   error
}

// Error returns error message of ErrorWithStack
func (e *ErrorWithStack) Error() string {
	return e.Err.Error()
}

// StackTrace implements IStack to be compatible with sentry
func (e *ErrorWithStack) StackTrace() errors.StackTrace {
	return e.Stack
}
