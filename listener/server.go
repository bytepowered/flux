package listener

import (
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/spf13/viper"
	"net/http"
)

type (
	Option func(server flux.WebListener)
)

func New(id string, config *flux.Configuration, wis []flux.WebInterceptor, opts ...Option) flux.WebListener {
	opts = append([]Option{
		WithErrorHandler(DefaultServerErrorHandler),
		WithNotfoundHandler(DefaultNotfoundHandler),
		WithResponseWriter(DefaultResponseWriter),
		WithInterceptors(wis),
	}, opts...)
	return NewWebListenerWith(id, config, opts...)
}

func NewWebListenerWith(id string, config *flux.Configuration, opts ...Option) flux.WebListener {
	if config == nil {
		config = flux.NewConfigurationOfViper(viper.New())
	}
	listener := ext.WebListenerFactory()(id, config)
	for _, opt := range opts {
		opt(listener)
	}
	return listener
}

func WithResponseWriter(writer flux.WebResponseWriter) Option {
	return func(server flux.WebListener) {
		server.SetResponseWriter(writer)
	}
}

func WithErrorHandler(handler flux.WebErrorHandler) Option {
	return func(server flux.WebListener) {
		server.SetErrorHandler(handler)
	}
}

func WithNotfoundHandler(f flux.WebHandler) Option {
	return func(server flux.WebListener) {
		server.SetNotfoundHandler(f)
	}
}

func WithInterceptors(array []flux.WebInterceptor) Option {
	return WithInterceptor(array...)
}

func WithInterceptor(array ...flux.WebInterceptor) Option {
	return func(server flux.WebListener) {
		for _, wi := range array {
			server.AddInterceptor(wi)
		}
	}
}

func WithWebHandlers(tuples []WebHandlerTuple) Option {
	return WithWebHandler(tuples...)
}

func WithWebHandler(tuples ...WebHandlerTuple) Option {
	return func(server flux.WebListener) {
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
	return func(server flux.WebListener) {
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
