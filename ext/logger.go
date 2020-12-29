package ext

import (
	"context"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
)

var (
	loggerFactory flux.LoggerFactory
)

func StoreLoggerFactory(f flux.LoggerFactory) {
	loggerFactory = pkg.RequireNotNil(f, "LoggerFactory is nil").(flux.LoggerFactory)
}

// NewLoggerWith
func NewLoggerWith(values context.Context) flux.Logger {
	return loggerFactory(values)
}

// NewLogger ...
func NewLogger() flux.Logger {
	return NewLoggerWith(context.TODO())
}
