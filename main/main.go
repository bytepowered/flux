package main

import (
	"errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
	"time"
)

import (
	_ "github.com/apache/dubbo-go/filter/filter_impl"
	_ "github.com/apache/dubbo-go/registry/zookeeper"
)

import (
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
	server.InitLogger()
	build := flux.Build{CommitId: GitCommit, Version: Version, Date: BuildDate}
	server.InitAppConfig(server.EnvKeyDeployEnv)
	dm := NewDispatcherManager()
	if err := dm.Prepare(); nil != err {
		logger.Panic("DispatcherManager prepare:", err)
	}
	if err := dm.Init(); nil != err {
		logger.Panic("DispatcherManager init:", err)
	}
	go func() {
		if err := dm.Startup(build); nil != err && !errors.Is(err, http.ErrServerClosed) {
			logger.Error(err)
		}
	}()
	quit := make(chan os.Signal, 1)
	dm.AwaitSignal(quit, 10*time.Second)
}

func NewDispatcherManager(options ...server.OptionFunc) *server.DispatcherManager {
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
