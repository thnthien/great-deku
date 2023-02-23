package trace

import "time"

// ISpan ...
type ISpan interface {
	NeedExport(b bool)
	IsRecordingEvents() bool
	SetAttribute(key string, val interface{})
	SetError(err error)
	IsError() bool
	GetSpanData() ISpanData
	GetTraceID() string
	GetSpanID() string
	End()
	EndExport()
	SetWarnDuration(d time.Duration)
}

type ISpanData interface {
	GetTraceID() string
	GetSpanID() string
}
