package logger

import (
	"github.com/bytepowered/flux"
)

var (
	_simLogger flux.Logger
)

// SetSimpleLogger set simple logger instance
func SetSimpleLogger(logger flux.Logger) {
	_simLogger = logger
}

// SimpleLogger get a simple logger instance
func SimpleLogger() flux.Logger {
	return _simLogger
}

// Info ...
func Info(args ...interface{}) {
	_simLogger.Info(args...)
}

// Warn ...
func Warn(args ...interface{}) {
	_simLogger.Warn(args...)
}

// Error ...
func Error(args ...interface{}) {
	_simLogger.Error(args...)
}

// Debug ...
func Debug(args ...interface{}) {
	_simLogger.Debug(args...)
}

// Debug ...
func Panic(args ...interface{}) {
	_simLogger.Panic(args...)
}

// Infof ...
func Infof(fmt string, args ...interface{}) {
	_simLogger.Infof(fmt, args...)
}

// Warnf ...
func Warnf(fmt string, args ...interface{}) {
	_simLogger.Warnf(fmt, args...)
}

// Errorf ...
func Errorf(fmt string, args ...interface{}) {
	_simLogger.Errorf(fmt, args...)
}

// Debugf ...
func Debugf(fmt string, args ...interface{}) {
	_simLogger.Debugf(fmt, args...)
}

// Debugf ...
func Panicf(fmt string, args ...interface{}) {
	_simLogger.Panicf(fmt, args...)
}

// Infof ...
func Infow(msg string, keysAndValues ...interface{}) {
	_simLogger.Infow(msg, keysAndValues...)
}

// Warnf ...
func Warnw(msg string, keysAndValues ...interface{}) {
	_simLogger.Warnw(msg, keysAndValues...)
}

// Errorf ...
func Errorw(msg string, keysAndValues ...interface{}) {
	_simLogger.Errorw(msg, keysAndValues...)
}

// Debugf ...
func Debugw(msg string, keysAndValues ...interface{}) {
	_simLogger.Debugw(msg, keysAndValues...)
}

// Debugf ...
func Panicw(msg string, keysAndValues ...interface{}) {
	_simLogger.Panicw(msg, keysAndValues...)
}
