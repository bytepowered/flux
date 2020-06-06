package server

import (
	"context"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
)

const (
	// 默认外部化配置的日志文件环境变量名称，与Dubbo的环境变量保持一致
	EnvKeyApplicationLogConfFile = "APP_LOG_CONF_FILE"
)

var (
	_defZapConfig = zap.NewProductionConfig()
)

func DefaultLoggerFactory(values context.Context) flux.Logger {
	sugar := NewZapLogger(&_defZapConfig)
	if traceId := values.Value(logger.TraceId); nil != traceId {
		sugar = sugar.With(zap.String(logger.TraceId, cast.ToString(traceId)))
	}
	return sugar
}

func LoadLoggerConfig(file string) (zap.Config, error) {
	config := zap.Config{}
	if file == "" {
		file = os.Getenv(EnvKeyApplicationLogConfFile)
	}
	if file == "" {
		return config, errors.New("log configure file name is nil")
	}
	if path.Ext(file) != ".yml" {
		return config, errors.Errorf("log configure file name{%s} suffix must be .yml", file)
	}
	confFileStream, err := ioutil.ReadFile(file)
	if err != nil {
		return config, errors.Errorf("ioutil.ReadFile(file:%s) = error:%v", file, err)
	}
	err = yaml.Unmarshal(confFileStream, &config)
	if err != nil {
		return config, errors.Errorf("[Unmarshal]init _defaultLogger error: %v", err)
	}
	return config, nil
}

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
