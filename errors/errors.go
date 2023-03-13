package errors

import (
	"context"
	"fmt"
	"runtime"
	"strconv"

	"github.com/pkg/errors"

	spancontext "github.com/thnthien/great-deku/trace/pkg/span-context"
)

const gray, resetColor = "\x1b[90m", "\x1b[0m"

type Code int

var httpStatusMessages = map[string]string{
	"100": "Continue",
	"101": "Switching Protocols",
	"102": "Processing (WebDAV)",
	"103": "Early Hints Experimental",
	"200": "OK",
	"201": "Created",
	"202": "Accepted",
	"203": "Non-Authoritative Information",
	"204": "No Content",
	"205": "Reset Content",
	"206": "Partial Content",
	"207": "Multi-Status (WebDAV)",
	"208": "Already Reported (WebDAV)",
	"226": "IM Used (HTTP Delta encoding)",
	"300": "Multiple Choices",
	"301": "Moved Permanently",
	"302": "Found",
	"303": "See Other",
	"304": "Not Modified",
	"305": "Use Proxy Deprecated",
	"306": "unused",
	"307": "Temporary Redirect",
	"308": "Permanent Redirect",
	"400": "Bad Request",
	"401": "Unauthorized",
	"402": "Payment Required Experimental",
	"403": "Forbidden",
	"404": "Not Found",
	"405": "Method Not Allowed",
	"406": "Not Acceptable",
	"407": "Proxy Authentication Required",
	"408": "Request Timeout",
	"409": "Conflict",
	"410": "Gone",
	"411": "Length Required",
	"412": "Precondition Failed",
	"413": "Payload Too Large",
	"414": "URI Too Long",
	"415": "Unsupported Media Type",
	"416": "Range Not Satisfiable",
	"417": "Expectation Failed",
	"418": "I'm a teapot",
	"421": "Misdirected Request",
	"422": "Unprocessable Entity (WebDAV)",
	"423": "Locked (WebDAV)",
	"424": "Failed Dependency (WebDAV)",
	"425": "Too Early Experimental",
	"426": "Upgrade Required",
	"428": "Precondition Required",
	"429": "Too Many Requests",
	"431": "Request Header Fields Too Large",
	"451": "Unavailable For Legal Reasons",
	"500": "Internal Server Error",
	"501": "Not Implemented",
	"502": "Bad Gateway",
	"503": "Service Unavailable",
	"504": "Gateway Timeout",
	"505": "HTTP Version Not Supported",
	"506": "Variant Also Negotiates",
	"507": "Insufficient Storage (WebDAV)",
	"508": "Loop Detected (WebDAV)",
	"510": "Not Extended",
	"511": "Network Authentication Required",
}

// CustomCode defines a custom error code
type CustomCode struct {
	StdCode        Code
	String         string
	DefaultMessage string
}

func (c Code) String() string {
	code := strconv.Itoa(int(c))
	if len(code) >= 3 {
		if s, ok := httpStatusMessages[code[:3]]; ok {
			return s
		}
	}
	return "Code(" + code + ")"
}

// Error returns APIError with provided information
func Error(code Code, message string, errs ...error) *APIError {
	return newError(false, code, message, errs...)
}

// ErrorTrace ...
func ErrorTrace(code Code, message string, errs ...error) *APIError {
	return newError(true, code, message, errs...)
}

// ErrorTraceCtx ...
func ErrorTraceCtx(ctx context.Context, code Code, message string, errs ...error) *APIError {
	xerr := newError(true, code, message, errs...)

	if p := spancontext.FromContext(ctx); p != nil {
		xerr.TraceID = p.GetTraceID()
		xerr.SpanID = p.GetSpanID()
	}

	return xerr
}

// DefaultErrorMessage returns default error message of provided Code
func DefaultErrorMessage(code Code) string {
	return code.String()
}

func newError(trace bool, code Code, message string, errs ...error) *APIError {
	if message == "" {
		message = code.String()
	}

	var err error
	if len(errs) > 0 {
		err = errs[0]
	}

	// Overwrite *Error
	if xerr, ok := err.(*APIError); ok && xerr != nil {
		// Keep original message
		if xerr.Original == "" {
			xerr.Original = xerr.Message
		}
		xerr.Code = code
		xerr.Message = message
		xerr.Trace = xerr.Trace || trace
		return xerr
	}

	// Wrap error with stacktrace
	return &APIError{
		Err:      err,
		Code:     code,
		Message:  message,
		Original: "",
		Stack:    errors.New("").(IStack).StackTrace(),
		Trace:    trace,
	}
}

func HandleRecover(handler func(error, string)) {
	// Size of the stack to be printed.
	var stackSize int = 4 << 10 // 4 KB
	if r := recover(); r != nil {
		err, ok := r.(error)
		if !ok {
			err = fmt.Errorf("%v", r)
		}
		stack := make([]byte, stackSize)
		length := runtime.Stack(stack, false)
		handler(err, string(stack[:length]))
		return
	}
	handler(nil, "")
}
