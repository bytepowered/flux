package server

import (
	"context"
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/ext"
	"github.com/bytepowered/flux/flux-node/logger"
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

func InitLogger() {
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

func Bootstrap(build flux.Build) {
	InitAppConfig(EnvKeyDeployEnv)
	server := NewDefaultBootstrapServer()
	if err := server.Prepare(); nil != err {
		logger.Panic("BootstrapServer prepare:", err)
	}
	if err := server.Initial(); nil != err {
		logger.Panic("BootstrapServer init:", err)
	}
	go func() {
		if err := server.Startup(build); nil != err && err != http.ErrServerClosed {
			logger.Error(err)
		}
	}()
	quit := make(chan os.Signal, 1)
	server.OnSignalShutdown(quit, 10*time.Second)
}

func IsDisabled(config *flux.Configuration) bool {
	return config.GetBool("disable") || config.GetBool("disabled")
}
