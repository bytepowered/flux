package ext

import (
	"github.com/bytepowered/flux"
)

var (
	_fluxLogger flux.Logger
)

// SetLogger ...
func SetLogger(logger flux.Logger) {
	_fluxLogger = logger
}

// GetLogger ...
func GetLogger() flux.Logger {
	return _fluxLogger
}

// SetLoggerLevel ...
func SetLoggerLevel(level string) bool {
	if l, ok := _fluxLogger.(OpsLogger); ok {
		l.SetLoggerLevel(level)
		return true
	}
	return false
}

// OpsLogger ...
type OpsLogger interface {
	flux.Logger
	SetLoggerLevel(level string)
}
