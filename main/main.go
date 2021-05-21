package main

import (
	"errors"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/listener"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/server"
	_ "github.com/bytepowered/flux/transporter/dubbo"
	_ "github.com/bytepowered/flux/transporter/echo"
	_ "github.com/bytepowered/flux/transporter/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
	"time"
)

import (
	_ "github.com/apache/dubbo-go/filter/filter_impl"
	_ "github.com/apache/dubbo-go/registry/zookeeper"
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
	generic := NewDefaultGenericServer()
	if err := generic.Prepare(); nil != err {
		logger.Panic("GenericServer prepare:", err)
	}
	if err := generic.Initial(); nil != err {
		logger.Panic("GenericServer init:", err)
	}
	go func() {
		if err := generic.Startup(build); nil != err && !errors.Is(err, http.ErrServerClosed) {
			logger.Error(err)
		}
	}()
	quit := make(chan os.Signal, 1)
	generic.AwaitSignal(quit, 10*time.Second)
}

func NewDefaultGenericServer(options ...server.GenericOptionFunc) *server.GenericServer {
	opts := []server.GenericOptionFunc{
		server.WithServerBanner("Flux.go"),
		// Lookup version
		server.WithVersionLookupFunc(func(webex flux.ServerWebContext) string {
			return webex.HeaderVar(server.DefaultHttpHeaderVersion)
		}),
		// Default WebListener
		server.WithWebListener(listener.New(server.ListenerIdDefault,
			server.NewWebListenerOptions(server.ListenerIdDefault), nil)),
		// Admin WebListener
		server.WithWebListener(listener.New(server.ListenServerIdAdmin,
			server.NewWebListenerOptions(server.ListenServerIdAdmin), nil,
			// 内部元数据查询
			listener.WithWebHandlers([]listener.WebHandlerTuple{
				// Metrics
				{Method: "GET", Pattern: "/inspect/metrics", Handler: flux.WrapHttpHandler(promhttp.Handler())},
			}),
		)),
	}
	return server.NewGenericServer(append(opts, options...)...)
}
