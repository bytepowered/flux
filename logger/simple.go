package logger

import (
	"github.com/bytepowered/flux"
	"go.uber.org/zap"
)

var (
	simLogger *zap.SugaredLogger
)

func init() {
	SetSimpleLogger(zap.S())
}

// SetSimpleLogger set simple logger instance
func SetSimpleLogger(logger *zap.SugaredLogger) {
	simLogger = logger.Desugar().WithOptions(zap.AddCallerSkip(1)).Sugar()
}

// SimpleLogger get a simple logger instance
func SimpleLogger() flux.Logger {
	return simLogger
}

// Info ...
func Info(args ...interface{}) {
	simLogger.Info(args...)
}

// Warn ...
func Warn(args ...interface{}) {
	simLogger.Warn(args...)
}

// Error ...
func Error(args ...interface{}) {
	simLogger.Error(args...)
}

// Debug ...
func Debug(args ...interface{}) {
	simLogger.Debug(args...)
}

// Debug ...
func Panic(args ...interface{}) {
	simLogger.Panic(args...)
}

// Infof ...
func Infof(fmt string, args ...interface{}) {
	simLogger.Infof(fmt, args...)
}

// Warnf ...
func Warnf(fmt string, args ...interface{}) {
	simLogger.Warnf(fmt, args...)
}

// Errorf ...
func Errorf(fmt string, args ...interface{}) {
	simLogger.Errorf(fmt, args...)
}

// Debugf ...
func Debugf(fmt string, args ...interface{}) {
	simLogger.Debugf(fmt, args...)
}

// Debugf ...
func Panicf(fmt string, args ...interface{}) {
	simLogger.Panicf(fmt, args...)
}

// Infof ...
func Infow(msg string, keysAndValues ...interface{}) {
	simLogger.Infow(msg, keysAndValues...)
}

// Warnf ...
func Warnw(msg string, keysAndValues ...interface{}) {
	simLogger.Warnw(msg, keysAndValues...)
}

// Errorf ...
func Errorw(msg string, keysAndValues ...interface{}) {
	simLogger.Errorw(msg, keysAndValues...)
}

// Debugf ...
func Debugw(msg string, keysAndValues ...interface{}) {
	simLogger.Debugw(msg, keysAndValues...)
}

// Debugf ...
func Panicw(msg string, keysAndValues ...interface{}) {
	simLogger.Panicw(msg, keysAndValues...)
}
