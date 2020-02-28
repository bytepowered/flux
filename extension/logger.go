package extension

import (
	"github.com/bytepowered/flux"
)

var (
	_logger flux.Logger
)

// SetLogger ...
func SetLogger(logger flux.Logger) {
	_logger = logger
}

// GetLogger ...
func GetLogger() flux.Logger {
	return _logger
}

// SetLoggerLevel ...
func SetLoggerLevel(level string) bool {
	if l, ok := _logger.(OpsLogger); ok {
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
