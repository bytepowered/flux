package logger

import "github.com/bytepowered/flux/ext"

// Info ...
func Info(args ...interface{}) {
	ext.GetLogger().Info(args...)
}

// Warn ...
func Warn(args ...interface{}) {
	ext.GetLogger().Warn(args...)
}

// Error ...
func Error(args ...interface{}) {
	ext.GetLogger().Error(args...)
}

// Debug ...
func Debug(args ...interface{}) {
	ext.GetLogger().Debug(args...)
}

// Debug ...
func Panic(args ...interface{}) {
	ext.GetLogger().Panic(args...)
}

// Infof ...
func Infof(fmt string, args ...interface{}) {
	ext.GetLogger().Infof(fmt, args...)
}

// Warnf ...
func Warnf(fmt string, args ...interface{}) {
	ext.GetLogger().Warnf(fmt, args...)
}

// Errorf ...
func Errorf(fmt string, args ...interface{}) {
	ext.GetLogger().Errorf(fmt, args...)
}

// Debugf ...
func Debugf(fmt string, args ...interface{}) {
	ext.GetLogger().Debugf(fmt, args...)
}

// Debugf ...
func Panicf(fmt string, args ...interface{}) {
	ext.GetLogger().Panicf(fmt, args...)
}
