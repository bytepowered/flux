package logger

import (
	"context"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
)

const (
	TraceId = "trace-id"
)

func Trace(traceId string) flux.Logger {
	return With(context.WithValue(context.Background(), TraceId, traceId))
}

func With(values context.Context) flux.Logger {
	return ext.NewLoggerWith(values)
}
