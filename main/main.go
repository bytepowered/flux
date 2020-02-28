package main

import (
	"github.com/bytepowered/flux/extension"
	"github.com/bytepowered/flux/internal"
)

var (
	GitCommit string
	Version   string
	BuildDate string
)

func main() {
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
	ver := internal.BuildVersion{CommitId: GitCommit, Version: Version, Date: BuildDate}
	logger.Error(app.Start(ver))
}
