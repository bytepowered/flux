package server

import (
	"context"
	dubgo "github.com/apache/dubbo-go/config"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"time"
)

const (
	EnvKeyDeployEnv = "DEPLOY_ENV"
)

func InitDefaultLogger() {
	config, err := LoadLoggerConfig("")
	if nil != err {
		panic(err)
	}
	sugar := NewZapLogger(config)
	logger.SetSimpleLogger(sugar)
	zap.ReplaceGlobals(sugar.Desugar())
	ext.StoreLoggerFactory(func(values context.Context) flux.Logger {
		if nil == values {
			return sugar
		}
		if traceId := values.Value(logger.TraceId); nil != traceId {
			return sugar.With(zap.String(logger.TraceId, cast.ToString(traceId)))
		}
		return sugar
	})
}

func InitConfiguration(envKey string) {
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
	InitConfiguration(EnvKeyDeployEnv)
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
	signal.Notify(quit, dubgo.ShutdownSignals...)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := engine.Shutdown(ctx); nil != err && err != http.ErrServerClosed {
		logger.Error(err)
	}
}
