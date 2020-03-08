package server

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/extension"
	"github.com/bytepowered/flux/internal"
	"github.com/bytepowered/flux/logger"
)

func InitDefaultLogger() {
	l, err := internal.InitLogger()
	if err != nil && l != nil {
		l.Panicf("FluxServer logger init: %v", err)
	} else {
		extension.SetLogger(l)
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
	if err := fx.Start(ver); nil != err {
		logger.Error(err)
	}
	fx.Shutdown()
}
