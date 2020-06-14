package server

import (
	"context"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"os"
	"reflect"
)

const (
	// 默认外部化配置的日志文件环境变量名称，与Dubbo的环境变量保持一致
	EnvKeyApplicationLogConfFile = "APP_LOG_CONF_FILE"
)

var (
	_defZapConfig = zap.NewProductionConfig()
	_defZapLogger = NewZapLogger(_defZapConfig)
)

func DefaultLoggerFactory(values context.Context) flux.Logger {
	newLogger := _defZapLogger
	if traceId := values.Value(logger.TraceId); nil != traceId {
		newLogger = newLogger.With(zap.String(logger.TraceId, cast.ToString(traceId)))
	}
	return newLogger
}

func LoadLoggerConfig(file string) (zap.Config, error) {
	if file == "" {
		file = os.Getenv(EnvKeyApplicationLogConfFile)
	}
	config := _defZapConfig
	if file == "" {
		return config, errors.New("log configure file name is nil")
	}
	v := viper.New()
	v.SetConfigFile(file)
	if err := v.ReadInConfig(); nil != err {
		return config, fmt.Errorf("read logger config, path: %s, err: %w", file, err)
	}
	// AtomicLevel转换
	if err := v.Unmarshal(&config, viper.DecodeHook(stringToAtomicLevel)); nil != err {
		return config, fmt.Errorf("unmarshal logger config, path: %s, err: %w", file, err)
	}
	return config, nil
}

func NewZapLogger(config zap.Config) *zap.SugaredLogger {
	zLogger, err := config.Build()
	if nil != err {
		panic(err)
	}
	return zLogger.Sugar()
}

func stringToAtomicLevel(f reflect.Kind, t reflect.Kind, data interface{}) (interface{}, error) {
	if f != reflect.String {
		return data, nil
	}
	if t != reflect.TypeOf(zap.AtomicLevel{}).Kind() {
		return data, nil
	}
	al := zap.NewAtomicLevel()
	return al, al.UnmarshalText([]byte(data.(string)))
}
