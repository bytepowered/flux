package server

import (
	"context"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"net/http"
	"os"
	"time"
)

const (
	EnvKeyDeployEnv = "DEPLOY_ENV"
)

func InitDefaultLogger() {
	config, err := logger.LoadConfig("")
	if nil != err {
		panic(err)
	}
	sugar := logger.NewZapLogger(config)
	logger.SetSimpleLogger(sugar)
	zap.ReplaceGlobals(sugar.Desugar())
	ext.SetLoggerFactory(func(values context.Context) flux.Logger {
		if nil == values {
			return sugar
		}
		if traceId := values.Value(logger.TraceId); nil != traceId {
			return sugar.With(zap.String(logger.TraceId, cast.ToString(traceId)))
		}
		return sugar
	})
}

func InitAppConfig(envKey string) {
	file := "application"
	env := os.Getenv(envKey)
	if env != "" {
		file = file + "-" + env
	}
	viper.SetConfigName(file)
	viper.AddConfigPath("/etc/flux/conf.d")
	viper.AddConfigPath("./conf.d")
	logger.Infof("Using config, file: %s, Env: %s", file, env)
	if err := viper.ReadInConfig(); nil != err {
		logger.Panicw("Fatal config error", "path", file, "error", err)
	}
}

func Run(ver flux.BuildInfo) {
	InitAppConfig(EnvKeyDeployEnv)
	engine := NewHttpServeEngine()
	if err := engine.Prepare(); nil != err {
		logger.Panic("HttpServeEngine prepare:", err)
	}
	if err := engine.Initial(); nil != err {
		logger.Panic("HttpServeEngine init:", err)
	}
	go func() {
		if err := engine.Startup(ver); nil != err && err != http.ErrServerClosed {
			logger.Error(err)
		}
	}()
	quit := make(chan os.Signal, 1)
	engine.OnSignalShutdown(quit, 10*time.Second)
}
