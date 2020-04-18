package server

import (
	"github.com/bytepowered/flux"
	"github.com/pkg/errors"
	"go.uber.org/zap"
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
type FluxLogger struct {
	flux.Logger
}

// InitLogger ...
func InitLogger() (flux.Logger, error) {
	file := os.Getenv(AppLogConfFile)
	if file == "" {
		return NewFluxLogger(nil), errors.New("log configure file name is nil")
	}
	if path.Ext(file) != ".yml" {
		return NewFluxLogger(nil), errors.Errorf("log configure file name{%s} suffix must be .yml", file)
	}
	confFileStream, err := ioutil.ReadFile(file)
	if err != nil {
		return NewFluxLogger(nil), errors.Errorf("ioutil.ReadFile(file:%s) = error:%v", file, err)
	}
	conf := &zap.Config{}
	err = yaml.Unmarshal(confFileStream, conf)
	if err != nil {
		return NewFluxLogger(nil), errors.Errorf("[Unmarshal]init _defaultLogger error: %v", err)
	}
	return NewFluxLogger(conf), nil
}

// initZapLogger ...
func NewFluxLogger(conf *zap.Config) flux.Logger {
	var zLogConfig zap.Config
	if conf != nil {
		zLogConfig = *conf
	} else {
		zLogConfig = zap.NewDevelopmentConfig()
	}
	zLogger, _ := zLogConfig.Build()
	return &FluxLogger{Logger: zLogger.Sugar()}
}
