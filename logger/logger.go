package logger

import "github.com/bytepowered/flux/extension"

// Info ...
func Info(args ...interface{}) {
	extension.GetLogger().Info(args...)
}

// Warn ...
func Warn(args ...interface{}) {
	extension.GetLogger().Warn(args...)
}

// Error ...
func Error(args ...interface{}) {
	extension.GetLogger().Error(args...)
}

// Debug ...
func Debug(args ...interface{}) {
	extension.GetLogger().Debug(args...)
}

// Debug ...
func Panic(args ...interface{}) {
	extension.GetLogger().Panic(args...)
}

// Infof ...
func Infof(fmt string, args ...interface{}) {
	extension.GetLogger().Infof(fmt, args...)
}

// Warnf ...
func Warnf(fmt string, args ...interface{}) {
	extension.GetLogger().Warnf(fmt, args...)
}

// Errorf ...
func Errorf(fmt string, args ...interface{}) {
	extension.GetLogger().Errorf(fmt, args...)
}

// Debugf ...
func Debugf(fmt string, args ...interface{}) {
	extension.GetLogger().Debugf(fmt, args...)
}

// Debugf ...
func Panicf(fmt string, args ...interface{}) {
	extension.GetLogger().Panicf(fmt, args...)
}
