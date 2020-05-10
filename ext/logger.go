package ext

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/pkg"
)

var (
	_fluxLogger flux.Logger
)

// SetLogger ...
func SetLogger(logger flux.Logger) {
	_fluxLogger = pkg.RequireNotNil(logger, "Logger is nil").(flux.Logger)
}

// GetLogger ...
func GetLogger() flux.Logger {
	return _fluxLogger
}
