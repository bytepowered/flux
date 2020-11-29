package ext

import (
	"context"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
)

var (
	_loggerFactory flux.LoggerFactory
)

func StoreLoggerFactory(f flux.LoggerFactory) {
	_loggerFactory = pkg.RequireNotNil(f, "LoggerFactory is nil").(flux.LoggerFactory)
}

// NewLoggerWith
func NewLoggerWith(values context.Context) flux.Logger {
	return _loggerFactory(values)
}

// NewLogger ...
func NewLogger() flux.Logger {
	return NewLoggerWith(context.TODO())
}
