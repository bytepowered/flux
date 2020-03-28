package server

import (
	"context"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/internal"
	"github.com/bytepowered/flux/logger"
	"os"
	"os/signal"
	"time"
)

func InitDefaultLogger() {
	l, err := internal.InitLogger()
	if err != nil && l != nil {
		l.Panicf("FluxServer logger init: %v", err)
	} else {
		ext.SetLogger(l)
	}
	if nil == l {
		panic("logger is nil")
	}
}

func Run(ver flux.BuildInfo) {
	fx := NewFluxServer()
	if err := fx.Init(LoadConfig()); nil != err {
		logger.Panicf("FluxServer init: %v", err)
	}
	go func() {
		if err := fx.Start(ver); nil != err {
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
