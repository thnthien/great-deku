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

const (
	// OK is returned on success.
	OK Code = 0

	Canceled Code = 1

	Unknown Code = 2

	InvalidArgument Code = 3

	DeadlineExceeded Code = 4

	NotFound Code = 5

	AlreadyExists Code = 6

	PermissionDenied Code = 7

	ResourceExhausted Code = 8

	FailedPrecondition Code = 9

	Aborted Code = 10

	OutOfRange Code = 11

	Unimplemented Code = 12

	Internal Code = 13

	Unavailable Code = 14

	DataLoss Code = 15

	Unauthenticated Code = 16

	WrongPassword = Code(1005)

	_maxCode = 17
)

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

var (
	mapCodes       [_maxCode]string
	mapCustomCodes map[Code]*CustomCode
)

func (c Code) String() string {
	if c >= 0 && int(c) < len(mapCodes) {
		return mapCodes[c]
	}
	if s := mapCustomCodes[c]; s != nil {
		return s.String
	}
	code := strconv.Itoa(int(c))
	if len(code) >= 3 {
		if s, ok := httpStatusMessages[code[:3]]; ok {
			return s
		}
	}
	return "Code(" + code + ")"
}

func init() {
	mapCodes[OK] = "OK"
	mapCodes[Canceled] = "Canceled"
	mapCodes[Unknown] = "Unknown"
	mapCodes[InvalidArgument] = "InvalidArgument"
	mapCodes[DeadlineExceeded] = "DeadlineExceeded"
	mapCodes[NotFound] = "NotFound"
	mapCodes[AlreadyExists] = "AlreadyExists"
	mapCodes[PermissionDenied] = "PermissionDenied"
	mapCodes[ResourceExhausted] = "ResourceExhausted"
	mapCodes[FailedPrecondition] = "FailedPrecondition"
	mapCodes[Aborted] = "Aborted"
	mapCodes[OutOfRange] = "OutOfRange"
	mapCodes[Unimplemented] = "OK"
	mapCodes[Internal] = "Internal"
	mapCodes[Unavailable] = "Unavailable"
	mapCodes[DataLoss] = "DataLoss"
	mapCodes[Unauthenticated] = "Unauthenticated"

	mapCustomCodes = make(map[Code]*CustomCode)
	mapCustomCodes[WrongPassword] = &CustomCode{Unauthenticated, "WRONG_PASSWORD", "Wrong password"}

}

// IsValidStandardErrorCode check if error code valid or not
func IsValidStandardErrorCode(c Code) bool {
	return c >= 0 && int(c) < len(mapCodes)
}

// GetCustomCode return CustomeCode object from Code
func GetCustomCode(c Code) *CustomCode {
	return mapCustomCodes[c]
}

// IsValidErrorCode check if Code is valid or not
func IsValidErrorCode(c Code) bool {
	return IsValidStandardErrorCode(c) || mapCustomCodes[c] != nil
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
	if code < _maxCode {
		return mapCodes[code]
	}
	if s := mapCustomCodes[code]; s != nil {
		return s.DefaultMessage
	}
	return "Unknown"
}

func newError(trace bool, code Code, message string, errs ...error) *APIError {
	if message == "" {
		message = DefaultErrorMessage(code)
	}

	var err error
	if len(errs) > 0 {
		err = errs[0]
	}

	var xcode Code
	if !IsValidStandardErrorCode(code) {
		xcode = code
		if c := mapCustomCodes[code]; c != nil {
			code = c.StdCode
		} else {
			code = Internal
		}
	}

	// Overwrite *Error
	if xerr, ok := err.(*APIError); ok && xerr != nil {
		// Keep original message
		if xerr.Original == "" {
			xerr.Original = xerr.Message
		}
		xerr.Code = code
		xerr.XCode = xcode
		xerr.Message = message
		xerr.Trace = xerr.Trace || trace
		return xerr
	}

	// Wrap error with stacktrace
	return &APIError{
		Err:      err,
		Code:     code,
		XCode:    xcode,
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
