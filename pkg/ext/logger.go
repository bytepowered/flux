package ext

import (
	"context"
	"github.com/bytepowered/fluxgo/pkg/flux"
)

type LoggerFactory func(values context.Context) flux.Logger

var (
	loggerFactory LoggerFactory
)

func SetLoggerFactory(f LoggerFactory) {
	flux.AssertNotNil(f, "LoggerFactory must not nil")
	loggerFactory = f
}

func NewLoggerWith(ctx context.Context) flux.Logger {
	return loggerFactory(ctx)
}

// NewLogger ...
func NewLogger() flux.Logger {
	return NewLoggerWith(context.TODO())
}
