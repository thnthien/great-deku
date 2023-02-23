package errors

import (
	"fmt"

	"go.uber.org/zap/zapcore"
)

// IError defines error interface returned by errors package
type IError interface {
	error
	IStack

	Format(st fmt.State, verb rune)
	Log(msg string, fields ...zapcore.Field) IError
}
