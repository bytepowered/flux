package bootstrap

import (
	"github.com/bytepowered/flux/extension"
	"github.com/bytepowered/flux/internal"
)

func Run(ver internal.BuildVersion) {
	//init logger
	logger, err := internal.InitLogger()
	if err != nil && logger != nil {
		logger.Panicf("Application logger init: %v", err)
	} else {
		extension.SetLogger(logger)
	}
	if nil == logger {
		panic("logger is nil")
	}
	extension.LoadConfig()
	app := internal.NewApplication()
	if err := app.Init(); nil != err {
		logger.Panicf("Application init: %v", err)
	}
	defer app.Shutdown()
	logger.Error(app.Start(ver))
}
