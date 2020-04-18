package server

import (
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

// InitLogger ...
func InitLogger() (*zap.SugaredLogger, error) {
	file := os.Getenv(AppLogConfFile)
	if file == "" {
		return NewZapLogger(nil), errors.New("log configure file name is nil")
	}
	if path.Ext(file) != ".yml" {
		return NewZapLogger(nil), errors.Errorf("log configure file name{%s} suffix must be .yml", file)
	}
	confFileStream, err := ioutil.ReadFile(file)
	if err != nil {
		return NewZapLogger(nil), errors.Errorf("ioutil.ReadFile(file:%s) = error:%v", file, err)
	}
	conf := &zap.Config{}
	err = yaml.Unmarshal(confFileStream, conf)
	if err != nil {
		return NewZapLogger(nil), errors.Errorf("[Unmarshal]init _defaultLogger error: %v", err)
	}
	return NewZapLogger(conf), nil
}

// initZapLogger ...
func NewZapLogger(conf *zap.Config) *zap.SugaredLogger {
	var zLogConfig zap.Config
	if conf != nil {
		zLogConfig = *conf
	} else {
		zLogConfig = zap.NewDevelopmentConfig()
	}
	zLogger, _ := zLogConfig.Build()
	return zLogger.Sugar()
}
