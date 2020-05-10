package server

import (
	"context"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"time"
)

const (
	EnvKeyDeployEnv = "DEPLOY_ENV"
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
		logger.Panicf("Fatal config error, path: %s, err: ", file, err)
	}
}

func Run(ver flux.BuildInfo) {
	InitConfiguration(EnvKeyDeployEnv)
	fx := NewFluxServer()
	if err := fx.Prepare(); nil != err {
		logger.Panic("FluxServer prepare:", err)
	}
	if err := fx.Initial(); nil != err {
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
