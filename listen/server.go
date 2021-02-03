package listen

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/webserver"
	"github.com/spf13/viper"
	"net/http"
)

type (
	Option func(server flux.ListenServer)
)

func NewServer(config *flux.Configuration, wis []flux.WebInterceptor, opts ...Option) flux.ListenServer {
	opts = append([]Option{
		WithServerErrorHandler(DefaultServerErrorHandler),
		WithNotfoundHandler(DefaultNotfoundHandler),
		WithResponseWriter(DefaultResponseWriter),
		WithInterceptors(append([]flux.WebInterceptor{webserver.NewRequestIdInterceptor()}, wis...)),
	}, opts...)
	return NewServerWith(config, opts...)
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

func WithInterceptors(array []flux.WebInterceptor) Option {
	return WithInterceptor(array...)
}

func WithInterceptor(array ...flux.WebInterceptor) Option {
	return func(server flux.ListenServer) {
		for _, wi := range array {
			server.AddInterceptor(wi)
		}
	}
}

func WithWebHandlers(tuples []WebHandlerTuple) Option {
	return WithWebHandler(tuples...)
}

func WithWebHandler(tuples ...WebHandlerTuple) Option {
	return func(server flux.ListenServer) {
		for _, h := range tuples {
			server.AddHandler(h.Method, h.Pattern, h.Handler)
		}
	}
}

type WebHandlerTuple struct {
	Method  string
	Pattern string
	Handler flux.WebHandler
}

func WithHttpHandlers(tuples []HttpHandlerTuple) Option {
	return WithHttpHandler(tuples...)
}

func WithHttpHandler(tuples ...HttpHandlerTuple) Option {
	return func(server flux.ListenServer) {
		for _, h := range tuples {
			server.AddHttpHandler(h.Method, h.Pattern, http.HandlerFunc(h.Handler))
		}
	}
}

type HttpHandlerTuple struct {
	Method  string
	Pattern string
	Handler func(http.ResponseWriter, *http.Request)
}
