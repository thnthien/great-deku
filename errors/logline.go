package errors

import (
	"encoding/json"
	"strconv"

	"go.uber.org/zap/zapcore"
)

// ValueOf  recieve input param is zapcore.Field and return a Interface
func ValueOf(f zapcore.Field) interface{} {
	switch {
	case f.Integer != 0:
		return f.Integer
	case f.String != "":
		return f.String
	}
	return f.Interface
}

// LogLine represents information of Logline
type LogLine struct {
	Level   string
	File    string
	Line    int
	Message string
	Fields  []zapcore.Field
}

// MarshalJSON returns log current with json formatted
func (l LogLine) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, 512)
	return l.MarshalTo(b), nil
}

// MarshalTo changes log current to json and assign to input byte array
func (l LogLine) MarshalTo(b []byte) []byte {
	b = append(b, '{')

	b = append(b, marshalLogLineField("@"+l.Level)...)
	b = append(b, ':')
	b = append(b, marshalLogLineField(l.Message)...)
	b = append(b, ',')

	b = append(b, `"@file":`...)
	b = append(b, marshalLogLineField(l.File+":"+strconv.Itoa(l.Line))...)

	for _, field := range l.Fields {
		b = append(b, ',')
		b = append(b, marshalLogLineField(field.Key)...)
		b = append(b, ':')

		if field.Integer != 0 {
			b = append(b, strconv.Itoa(int(field.Integer))...)
		} else if field.String != "" {
			b = append(b, marshalLogLineField(field.String)...)
		} else {
			b = append(b, marshalLogLineField(field.Interface)...)
		}
	}
	b = append(b, '}')
	return b
}

func marshalLogLineField(v interface{}) []byte {
	if xerr, ok := v.(*APIError); ok {
		data, _ := xerr.MarshalJSON()
		return data
	}
	data, err := json.Marshal(v)
	if err != nil {
		data, _ = json.Marshal(err)
	}
	return data
}
