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

// 注意：自定义实现main方法时，需要导入WebServer实现模块；
// 或者导入 _ "github.com/bytepowered/flux/webecho" 自动注册WebServer；
func main() {
	server.InitLogger()
	Bootstrap(flux.Build{CommitId: GitCommit, Version: Version, Date: BuildDate})
}

func Bootstrap(build flux.Build) {
	server.InitAppConfig(server.EnvKeyDeployEnv)
	bootstrap := NewDefaultBootstrapServer()
	if err := bootstrap.Prepare(); nil != err {
		logger.Panic("BootstrapServer prepare:", err)
	}
	if err := bootstrap.Initial(); nil != err {
		logger.Panic("BootstrapServer init:", err)
	}
	go func() {
		if err := bootstrap.Startup(build); nil != err && !errors.Is(err, http.ErrServerClosed) {
			logger.Error(err)
		}
	}()
	quit := make(chan os.Signal, 1)
	bootstrap.OnSignalShutdown(quit, 10*time.Second)
}

func NewDefaultBootstrapServer(options ...server.Option) *server.BootstrapServer {
	opts := []server.Option{
		server.WithServerBanner("Flux.go"),
		// Lookup version
		server.WithVersionLookupFunc(func(webex flux.ServerWebContext) string {
			return webex.HeaderVar(server.DefaultHttpHeaderVersion)
		}),
		// Default WebListener
		server.WithWebListener(listener.New(server.ListenerIdDefault, server.LoadWebListenerConfig(server.ListenerIdDefault), nil)),
		// Admin WebListener
		server.WithWebListener(listener.New(server.ListenServerIdAdmin, server.LoadWebListenerConfig(server.ListenServerIdAdmin), nil,
			// 内部元数据查询
			listener.WithWebHandlers([]listener.WebHandlerTuple{
				// Metrics
				{Method: "GET", Pattern: "/inspect/metrics", Handler: flux.WrapHttpHandler(promhttp.Handler())},
			}),
		)),
	}
	return server.NewBootstrapServerWith(append(opts, options...)...)
}
