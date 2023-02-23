package tracepkg

import (
	"encoding/json"

	"github.com/thnthien/great-deku/l"
)

type LogExporter struct {
	ll l.Logger
}

func (e *LogExporter) ExportSpan(sd *SpanData) {
	//go func() {
	str, _ := json.Marshal(sd)
	if sd.Error != nil {
		e.ll.Error(string(str))
		return
	}
	if sd.WarnDuration.Milliseconds() > 0 && sd.DurationVal.Milliseconds() > sd.WarnDuration.Milliseconds() {
		e.ll.Warn(string(str))
		return
	}
	if sd.MustLog {
		e.ll.Info(string(str))
		return
	}

	e.ll.Debug(string(str))
	//}()
}

func (e *LogExporter) Start() {
	RegisterExporter(e)
}

func NewLogExporter(ll l.Logger) *LogExporter {
	return &LogExporter{
		ll,
	}
}
