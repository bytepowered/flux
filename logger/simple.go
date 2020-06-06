package logger

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
)

var (
	_simLogger flux.Logger
)

func InitSimpleLogger() {
	_simLogger = ext.NewLogger()
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
