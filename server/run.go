package server

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/extension"
	"github.com/bytepowered/flux/internal"
)

func Run(ver flux.BuildInfo) {
	//init logger
	logger, err := internal.InitLogger()
	if err != nil && logger != nil {
		logger.Panicf("FluxServer logger init: %v", err)
	} else {
		extension.SetLogger(logger)
	}
	if nil == logger {
		panic("logger is nil")
	}
	app := NewFluxServer()
	if err := app.Init(extension.LoadConfig()); nil != err {
		logger.Panicf("FluxServer init: %v", err)
	}
	defer app.Shutdown()
	logger.Error(app.Start(ver))
}
