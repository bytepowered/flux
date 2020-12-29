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
	defaultZapConfig = zap.NewProductionConfig()
	defaultZapLogger = NewZapLogger(defaultZapConfig)
)

func DefaultLoggerFactory(values context.Context) flux.Logger {
	return SugaredLoggerFactoryFactory(defaultZapLogger)(values)
}

func SugaredLoggerFactoryFactory(sugar *zap.SugaredLogger) flux.LoggerFactory {
	return func(values context.Context) flux.Logger {
		if values == nil {
			return sugar
		}
		newLogger := sugar
		if traceId := values.Value(logger.TraceId); nil != traceId {
			newLogger = newLogger.With(zap.String(logger.TraceId, cast.ToString(traceId)))
		}
		if extras, ok := values.Value(logger.Extras).(map[string]string); ok && len(extras) > 0 {
			fields := make([]interface{}, 0, len(extras))
			for name, val := range extras {
				fields = append(fields, zap.String(name, val))
			}
			newLogger = newLogger.With(fields...)
		}
		return newLogger
	}
}

func LoadLoggerConfig(file string) (zap.Config, error) {
	if file == "" {
		file = os.Getenv(EnvKeyApplicationLogConfFile)
	}
	config := defaultZapConfig
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
