package errors

import (
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"strconv"

	"github.com/pkg/errors"
	"go.uber.org/zap/zapcore"
)

// APIError defines struct of error for API
type APIError struct {
	Code     Code
	Err      error
	Message  string
	Original string
	Stack    errors.StackTrace
	Trace    bool
	Trivial  bool
	Logs     []LogLine
	Meta     map[string]string
	TraceID  string
	SpanID   string
}

// Log appends log to Logs field of APIError
func (e *APIError) Log(msg string, fields ...zapcore.Field) IError {
	_, file, line, _ := runtime.Caller(1)
	e.Logs = append(e.Logs, LogLine{
		Level:   "error",
		File:    file,
		Line:    line,
		Fields:  fields,
		Message: msg,
	})
	return e
}

// Error return error message
func (e *APIError) Error() string {
	return e.Message
}

// Cause implemnts error and errors.StackTrace to be compatible with sentry
func (e *APIError) Cause() error {
	err := e.Err
	if err == nil {
		err = e
	}
	return &ErrorWithStack{
		Err:   err,
		Stack: e.Stack,
	}
}

// StackTrace returns Stack of APIError
func (e *APIError) StackTrace() errors.StackTrace {
	return e.Stack
}

// Format parse APIError to string with suitable format
// then write it to provided writer
func (e *APIError) Format(st fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case st.Flag('#') || st.Flag('+'):
			_, _ = fmt.Fprintf(st, "\ncode=%v message=%v", e.Code, e.Message)
			if e.Original != "" {
				_, _ = fmt.Fprintf(st, " original=%s", e.Original)
			}
			if e.Err != nil {
				_, _ = fmt.Fprintf(st, " cause=%+v", e.Err)
			}
			for k, v := range e.Meta {
				_, _ = fmt.Fprint(st, " ", k, "=", v)
			}
			for _, log := range e.Logs {
				_, _ = fmt.Fprint(st, "\n\t", log.Line, " ", gray, log.File, ":", strconv.Itoa(log.Line), resetColor)
				for k, v := range log.Fields {
					_, _ = fmt.Fprint(st, " ", k, "=", ValueOf(v))
				}
			}
			fallthrough
		case st.Flag('+'):
			_, _ = fmt.Fprintf(st, "%+v", e.StackTrace())
		default:
			_, _ = io.WriteString(st, e.Error())
		}
	case 's':
		_, _ = io.WriteString(st, e.Error())
	case 'q':
		_, _ = fmt.Fprintf(st, "%q", e.Error())
	}
}

// MarshalJSON jsonize APIError to bytes
func (e *APIError) MarshalJSON() ([]byte, error) {
	if e == nil {
		return []byte("null"), nil
	}

	b := make([]byte, 0, 2048)

	b = append(b, '{')
	b = append(b, `"code":`...)
	b = append(b, marshal(e.Code.String())...)

	if e.Err != nil {
		b = append(b, ',')
		b = append(b, `"err":`...)
		b = append(b, marshal(e.Err.Error())...)
	}

	b = append(b, ',')
	b = append(b, `"msg":`...)
	b = append(b, marshal(e.Message)...)

	if e.Original != "" {
		b = append(b, ',')
		b = append(b, `"orig":`...)
		b = append(b, marshal(e.Original)...)
	}

	b = append(b, ',')
	b = append(b, `"logs":`...)
	b = append(b, '[')
	for i, line := range e.Logs {
		if i > 0 {
			b = append(b, ',')
		}
		b = line.MarshalTo(b)
	}
	b = append(b, ']')

	if e.Trace {
		b = append(b, ',')
		b = append(b, `"stack":`...)
		b = append(b, marshal(fmt.Sprintf("%+v", e.Stack))...)
	}

	b = append(b, '}')
	return b, nil
}

func marshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		data, _ = json.Marshal(err)
	}
	return data
}
