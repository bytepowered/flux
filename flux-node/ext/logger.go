package ext

import (
	"context"
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-pkg"
)

var (
	loggerFactory flux.LoggerFactory
)

func SetLoggerFactory(f flux.LoggerFactory) {
	loggerFactory = fluxpkg.MustNotNil(f, "LoggerFactory is nil").(flux.LoggerFactory)
}

// NewLoggerWith
func NewLoggerWith(ctx context.Context) flux.Logger {
	return loggerFactory(ctx)
}

// NewLogger ...
func NewLogger() flux.Logger {
	return NewLoggerWith(context.TODO())
}
