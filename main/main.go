package main

import (
	"errors"
	"net/http"
	"os"
	"time"
)

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/cli/v2"
)

import (
	"github.com/bytepowered/fluxgo/pkg/cmd"
	"github.com/bytepowered/fluxgo/pkg/flux"
	"github.com/bytepowered/fluxgo/pkg/listener"
	"github.com/bytepowered/fluxgo/pkg/logger"
	"github.com/bytepowered/fluxgo/pkg/server"
	_ "github.com/bytepowered/fluxgo/pkg/transporter/dubbo"
	_ "github.com/bytepowered/fluxgo/pkg/transporter/echo"
	_ "github.com/bytepowered/fluxgo/pkg/transporter/http"
)

var (
	GitCommit string
	Version   string
	BuildDate string
)

func main() {
	build := flux.Build{CommitId: GitCommit, Version: Version, Date: BuildDate}
	app := cmd.NewApp(cmd.NewActions(
		cmd.InitLoggerAction,
		cmd.InitConfigAction,
		func(context *cli.Context) error {
			dm := newDispatcherManager()
			if err := dm.Prepare(); nil != err {
				return err
			}
			if err := dm.Init(); nil != err {
				return err
			}
			go func() {
				err := dm.Startup(build)
				if nil != err && !errors.Is(err, http.ErrServerClosed) {
					logger.Error(err)
				}
			}()
			quit := make(chan os.Signal, 1)
			dm.AwaitSignal(quit, 10*time.Second)
			return nil
		},
	), build)
	err := app.Run(os.Args)
	if err != nil {
		logger.Error(err)
	}
}

func newDispatcherManager(options ...server.OptionFunc) *server.DispatcherManager {
	opts := []server.OptionFunc{
		server.WithServerBanner("Flux.go"),
		// Default WebListener
		server.WithNewWebListener(listener.New(server.ListenerIdDefault,
			server.NewWebListenerOptions(server.ListenerIdDefault), nil)),
		// Admin WebListener
		server.WithNewWebListener(listener.New(server.ListenerIdAdmin,
			server.NewWebListenerOptions(server.ListenerIdAdmin), nil,
			// 内部元数据查询
			listener.WithHandlers([]listener.WebHandlerTuple{
				// Metrics
				{Method: "GET", Pattern: "/inspect/metrics", Handler: flux.WrapHttpHandler(promhttp.Handler())},
			}),
		)),
		// Setup
		server.EnabledRequestVersionLocator(server.ListenerIdDefault, server.DefaultRequestVersionLocateFunc),
		server.EnabledRequestVersionLocator(server.ListenerIdAdmin, server.DefaultRequestVersionLocateFunc),
	}
	return server.NewDispatcherManager(append(opts, options...)...)
}
