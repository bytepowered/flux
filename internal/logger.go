package internal

import (
	"github.com/bytepowered/flux"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
)

const (
	// 日志文件环境变量名称
	AppLogConfFile = "APP_LOG_CONF_FILE"
)

// Logger ...
type fluxLogger struct {
	flux.Logger
	dynamicLevel zap.AtomicLevel
}

// InitLogger ...
func InitLogger() (flux.Logger, error) {
	file := os.Getenv(AppLogConfFile)
	if file == "" {
		return initZapLogger(nil), errors.New("log configure file name is nil")
	}
	if path.Ext(file) != ".yml" {
		return initZapLogger(nil), errors.Errorf("log configure file name{%s} suffix must be .yml", file)
	}
	confFileStream, err := ioutil.ReadFile(file)
	if err != nil {
		return initZapLogger(nil), errors.Errorf("ioutil.ReadFile(file:%s) = error:%v", file, err)
	}

	conf := &zap.Config{}
	err = yaml.Unmarshal(confFileStream, conf)
	if err != nil {
		return initZapLogger(nil), errors.Errorf("[Unmarshal]init _defaultLogger error: %v", err)
	}
	return initZapLogger(conf), nil
}

// initZapLogger ...
func initZapLogger(conf *zap.Config) flux.Logger {
	var zapLoggerConfig zap.Config
	if conf == nil {
		zapLoggerConfig = zap.NewDevelopmentConfig()
		zapLoggerEncoderConfig := zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "_defaultLogger",
			CallerKey:      "caller",
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			EncodeLevel:    zapcore.CapitalColorLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}
		zapLoggerConfig.EncoderConfig = zapLoggerEncoderConfig
	} else {
		zapLoggerConfig = *conf
	}
	zapLogger, _ := zapLoggerConfig.Build(zap.AddCallerSkip(1))
	return &fluxLogger{Logger: zapLogger.Sugar(), dynamicLevel: zapLoggerConfig.Level}
}

// OpsLogger ...
type OpsLogger interface {
	flux.Logger
	SetLoggerLevel(level string)
}

// SetLoggerLevel ...
func (dl *fluxLogger) SetLoggerLevel(level string) {
	l := new(zapcore.Level)
	l.Set(level)
	dl.dynamicLevel.SetLevel(*l)
}
