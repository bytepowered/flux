package server

import (
	"context"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"os"
	"os/signal"
	"strings"
	"time"
)

const (
	templateEnvApplicationConfigPath = "conf.d/application-{env}.toml"
)

func InitDefaultLogger() {
	zLogger, err := InitLoggerDefault()
	if err != nil && zLogger != nil {
		zLogger.Panic("FluxServer logger init:", err)
	} else {
		ext.SetLogger(zLogger)
	}
	if nil == zLogger {
		panic("logger is nil")
	}
}

// 根据环境配置参数，获取带环境参数后缀的配置文件路径
func GetEnvBasedConfigPath() string {
	env := os.Getenv("env")
	if env == "" {
		env = os.Getenv("runtime.env")
	}
	if env != "" {
		path := strings.Replace(templateEnvApplicationConfigPath, "{env}", env, 1)
		// 如果环境后缀的配置文件不存在，返回空路径
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return ""
		} else {
			return path
		}
	} else {
		return ""
	}
}

func Run(ver flux.BuildInfo) {
	fx := NewFluxServer()
	globals := LoadConfig(GetEnvBasedConfigPath())
	if err := fx.Prepare(globals); nil != err {
		logger.Panic("FluxServer prepare:", err)
	}
	if err := fx.Init(globals); nil != err {
		logger.Panic("FluxServer init:", err)
	}
	go func() {
		if err := fx.Startup(ver); nil != err {
			logger.Error(err)
		}
	}()
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := fx.Shutdown(ctx); nil != err {
		logger.Error(err)
	}
}
