package ext

import (
	"context"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/toolkit"
)

var (
	loggerFactory flux.LoggerFactory
)

func SetLoggerFactory(f flux.LoggerFactory) {
	toolkit.AssertNotNil(f, "LoggerFactory must not nil")
	loggerFactory = f
}

// NewLoggerWith
func NewLoggerWith(ctx context.Context) flux.Logger {
	return loggerFactory(ctx)
}

// NewLogger ...
func NewLogger() flux.Logger {
	return NewLoggerWith(context.TODO())
}
