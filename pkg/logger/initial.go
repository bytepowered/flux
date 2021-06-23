package logger

import (
	"context"
	"fmt"
	"github.com/bytepowered/fluxgo/pkg/ext"
	"github.com/bytepowered/fluxgo/pkg/flux"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"os"
	"reflect"
	"runtime"
	"syscall"
)

const (
	// 默认外部化配置的日志文件环境变量名称，与Dubbo的环境变量保持一致
	EnvKeyApplicationLogConfFile = "APP_LOG_CONF_FILE"
)

var (
	defaultZapConfig = zap.NewProductionConfig()
	defaultZapLogger = NewZapLogger(defaultZapConfig)
)

func DefaultFactory(values context.Context) flux.Logger {
	return SugaredFactory(defaultZapLogger)(values)
}

func SugaredFactory(sugar *zap.SugaredLogger) ext.LoggerFactory {
	return func(values context.Context) flux.Logger {
		if values == nil {
			return sugar
		}
		newLogger := sugar
		if traceId := values.Value(TraceId); nil != traceId {
			newLogger = newLogger.With(zap.String(TraceId, cast.ToString(traceId)))
		}
		if extras, ok := values.Value(Extras).(map[string]string); ok && len(extras) > 0 {
			fields := make([]interface{}, 0, len(extras))
			for name, val := range extras {
				fields = append(fields, zap.String(name, val))
			}
			newLogger = newLogger.With(fields...)
		}
		return newLogger
	}
}

func LoadConfig(logfile string) (zap.Config, error) {
	if logfile == "" {
		logfile = os.Getenv(EnvKeyApplicationLogConfFile)
	}
	config := defaultZapConfig
	if logfile == "" {
		return config, errors.New("log configure logfile name is nil")
	}
	v := viper.New()
	v.SetConfigFile(logfile)
	if err := v.ReadInConfig(); nil != err {
		return config, fmt.Errorf("read logger config, path: %s, err: %w", logfile, err)
	}
	// AtomicLevel转换
	if err := v.Unmarshal(&config, viper.DecodeHook(stringToAtomicLevel)); nil != err {
		return config, fmt.Errorf("unmarshal logger config, path: %s, err: %w", logfile, err)
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

var (
	gStdErrFileHandler *os.File
)

func RewriteError(errfile string) error {
	file, err := os.OpenFile(errfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
		return err
	}
	gStdErrFileHandler = file
	if err = syscall.Dup2(int(file.Fd()), int(os.Stderr.Fd())); err != nil {
		fmt.Println(err)
		return err
	}
	// 内存回收前关闭文件描述符
	runtime.SetFinalizer(gStdErrFileHandler, func(fd *os.File) {
		_ = fd.Close()
	})
	return nil
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
