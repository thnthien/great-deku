package errors

import "github.com/pkg/errors"

// IStack defines an interface of Stack trace log
type IStack interface {
	StackTrace() errors.StackTrace
}
