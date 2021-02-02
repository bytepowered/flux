package listen

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/webserver"
	"github.com/spf13/viper"
)

type (
	Option func(server flux.ListenServer)
)

func NewServer(config *flux.Configuration, wis []flux.WebInterceptor) flux.ListenServer {
	return NewServerWith(config,
		WithServerErrorHandler(DefaultServerErrorHandler),
		WithNotfoundHandler(DefaultNotfoundHandler),
		WithResponseWriter(DefaultResponseWriter),
		WithInterceptors(append([]flux.WebInterceptor{
			webserver.NewRequestIdInterceptor(),
		}, wis...)),
	)
}

func NewServerWith(config *flux.Configuration, opts ...Option) flux.ListenServer {
	if config == nil {
		config = flux.NewConfiguration(viper.New())
	}
	server := ext.GetWebServerFactory()(config)
	for _, opt := range opts {
		opt(server)
	}
	return server
}

func WithResponseWriter(writer flux.WebResponseWriter) Option {
	return func(server flux.ListenServer) {
		server.SetResponseWriter(writer)
	}
}

func WithServerErrorHandler(handler flux.WebServerErrorHandler) Option {
	return func(server flux.ListenServer) {
		server.SetServerErrorHandler(handler)
	}
}

func WithNotfoundHandler(f flux.WebHandler) Option {
	return func(server flux.ListenServer) {
		server.SetNotfoundHandler(f)
	}
}

func WithInterceptors(wis []flux.WebInterceptor) Option {
	return func(server flux.ListenServer) {
		for _, wi := range wis {
			server.AddInterceptor(wi)
		}
	}
}
