package tracepkg

import (
	"context"
	t "runtime/trace"
	"sync"
	"sync/atomic"
	"time"

	"github.com/thnthien/great-deku/trace"
	"github.com/thnthien/great-deku/trace/pkg/id"
	spancontext "github.com/thnthien/great-deku/trace/pkg/span-context"
)

type Exporter interface {
	ExportSpan(s *SpanData)
}

type exportersMap map[Exporter]struct{}

var (
	exporterMu sync.Mutex
	exporters  atomic.Value
)

func RegisterExporter(e Exporter) {
	exporterMu.Lock()
	new := make(exportersMap)
	if old, ok := exporters.Load().(exportersMap); ok {
		for k, v := range old {
			new[k] = v
		}
	}
	new[e] = struct{}{}
	exporters.Store(new)
	exporterMu.Unlock()
}

// SpanData contains all the information collected by a Span.
type SpanData struct {
	Name string
	spancontext.SpanContext
	ParentSpanID id.SpanID
	StartTime    time.Time
	// The wall clock time of EndTime will be adjusted to always be offset
	// from StartTime by the duration of the span.
	EndTime      time.Time
	Duration     string
	DurationVal  time.Duration `json:"-"`
	WarnDuration time.Duration `json:"-"`
	// The values of Attributes each have type string, bool, or int64.
	Attributes map[string]interface{}

	Error interface{} `json:"error,omitempty"`
	// ChildSpanCount holds the number of child span created for this span.
	ChildSpanCount int
	MustLog        bool `json:"-"`
}

func (s SpanData) GetSpanID() string {
	return string(s.SpanID)
}

func (s SpanData) GetTraceID() string {
	return string(s.TraceID)
}

type Span struct {
	// data contains information recorded about the span.
	//
	// It will be non-nil if we are exporting the span or recording events for it.
	// Otherwise, data is nil, and the Span is simply a carrier for the
	// SpanContext, so that the trace ID is propagated.
	data        *SpanData
	mu          sync.Mutex // protects the contents of *data (but not the pointer value.)
	spanContext spancontext.SpanContext

	endOnce sync.Once

	executionTracerTaskEnd func() // ends the execution tracer span
	isExport               bool
}

func (s *Span) NeedExport(b bool) {
	s.isExport = b
}

// IsRecordingEvents returns true if events are being recorded for this span.
// Use this check to avoid computing expensive annotations when they will never
// be used.
func (s *Span) IsRecordingEvents() bool {
	if s == nil {
		return false
	}
	return s.data != nil
}

func (s *Span) addChild() {
	if !s.IsRecordingEvents() {
		return
	}
	s.mu.Lock()
	s.data.ChildSpanCount++
	s.mu.Unlock()
}

func (s *Span) SetAttribute(key string, val interface{}) {
	if s.data.Attributes == nil {
		s.data.Attributes = make(map[string]interface{})
	}
	s.data.Attributes[key] = val
}

func (s *Span) SetError(err error) {
	s.isExport = true
	s.data.Error = err
}
func (s *Span) IsError() bool {
	return s.data.Error != nil
}

func (s *Span) SetWarnDuration(d time.Duration) {
	s.data.WarnDuration = d
}

func (s *Span) GetSpanData() trace.ISpanData {
	return s.data
}

func (s *Span) GetTraceID() string {
	return string(s.data.TraceID)
}
func (s *Span) GetSpanID() string {
	return string(s.data.SpanID)
}

func StartSpan(ctx context.Context, name string) (context.Context, trace.ISpan) {
	var parent spancontext.SpanContext
	if p := spancontext.FromContext(ctx); p != nil {
		p.(*Span).addChild()
		parent = p.(*Span).spanContext
	}
	span := startSpanInternal(name, parent != spancontext.SpanContext{}, parent)
	ctx, end := startExecutionTracerTask(ctx, name)
	span.executionTracerTaskEnd = end
	return spancontext.NewContext(ctx, span), span
}

// End ends the span.
func (s *Span) End() {
	if s == nil {
		return
	}
	if s.executionTracerTaskEnd != nil {
		s.executionTracerTaskEnd()
	}
	if !s.IsRecordingEvents() {
		return
	}
	s.endOnce.Do(func() {
		sd := s.data
		sd.EndTime = MonotonicEndTime(sd.StartTime)
		sd.DurationVal = time.Since(sd.StartTime)
		sd.Duration = sd.DurationVal.String()

		//if s.isExport {
		exp, _ := exporters.Load().(exportersMap)
		mustExport := len(exp) > 0
		if mustExport {
			for e := range exp {
				e.ExportSpan(sd)
			}
		}
		//}
		//str, _ := json.Marshal(sd)
		//fmt.Printf("==> end span => endOnce %s\n", str)
	})
}

func (s *Span) EndExport() {
	if s == nil {
		return
	}
	if s.executionTracerTaskEnd != nil {
		s.executionTracerTaskEnd()
	}
	if !s.IsRecordingEvents() {
		return
	}
	s.endOnce.Do(func() {
		sd := s.data
		sd.EndTime = MonotonicEndTime(sd.StartTime)
		sd.DurationVal = time.Since(sd.StartTime)
		sd.Duration = sd.DurationVal.String()
		sd.MustLog = true

		exp, _ := exporters.Load().(exportersMap)
		mustExport := len(exp) > 0
		if mustExport {
			for e := range exp {
				e.ExportSpan(sd)
			}
		}

		// str, _ := json.Marshal(sd)
		// fmt.Printf("==> end span => endOnce %s\n", str)
	})
}

func MonotonicEndTime(start time.Time) time.Time {
	return start.Add(time.Since(start))
}

func startExecutionTracerTask(ctx context.Context, name string) (context.Context, func()) {
	if !t.IsEnabled() {
		// Avoid additional overhead if
		// runtime/trace is not enabled.
		return ctx, func() {}
	}
	nctx, task := t.NewTask(ctx, name)
	return nctx, task.End
}

func startSpanInternal(name string, hasParent bool, parent spancontext.SpanContext) *Span {
	span := &Span{}
	span.spanContext = parent

	if !hasParent {
		span.spanContext.TraceID = id.TraceID(id.TraceGen.NewTraceID())
	}
	span.spanContext.SpanID = id.SpanID(id.TraceGen.NewSpanID())

	span.data = &SpanData{
		SpanContext: span.spanContext,
		StartTime:   time.Now(),
		//SpanKind:        o.SpanKind,
		Name: name,
	}

	if hasParent {
		span.data.ParentSpanID = parent.SpanID
	}

	return span
}
