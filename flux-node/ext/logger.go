package ext

import (
	"context"
	flux2 "github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-pkg"
)

var (
	loggerFactory flux2.LoggerFactory
)

func SetLoggerFactory(f flux2.LoggerFactory) {
	loggerFactory = fluxpkg.MustNotNil(f, "LoggerFactory is nil").(flux2.LoggerFactory)
}

// NewLoggerWith
func NewLoggerWith(values context.Context) flux2.Logger {
	return loggerFactory(values)
}

// NewLogger ...
func NewLogger() flux2.Logger {
	return NewLoggerWith(context.TODO())
}
