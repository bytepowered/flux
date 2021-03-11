package listener

import (
	flux2 "github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/ext"
	"github.com/spf13/viper"
	"net/http"
)

type (
	Option func(server flux2.WebListener)
)

func New(id string, config *flux2.Configuration, wis []flux2.WebInterceptor, opts ...Option) flux2.WebListener {
	opts = append([]Option{
		WithErrorHandler(DefaultErrorHandler),
		WithNotfoundHandler(DefaultNotfoundHandler),
		WithResponseWriter(DefaultResponseWriter),
		WithInterceptors(wis),
	}, opts...)
	return NewWebListenerWith(id, config, opts...)
}

func NewWebListenerWith(id string, config *flux2.Configuration, opts ...Option) flux2.WebListener {
	if config == nil {
		config = flux2.NewConfigurationOfViper(viper.New())
	}
	listener := ext.WebListenerFactory()(id, config)
	for _, opt := range opts {
		opt(listener)
	}
	return listener
}

func WithResponseWriter(writer flux2.WebResponseWriter) Option {
	return func(server flux2.WebListener) {
		server.SetResponseWriter(writer)
	}
}

func WithErrorHandler(handler flux2.WebErrorHandler) Option {
	return func(server flux2.WebListener) {
		server.SetErrorHandler(handler)
	}
}

func WithNotfoundHandler(f flux2.WebHandler) Option {
	return func(server flux2.WebListener) {
		server.SetNotfoundHandler(f)
	}
}

func WithInterceptors(array []flux2.WebInterceptor) Option {
	return WithInterceptor(array...)
}

func WithInterceptor(array ...flux2.WebInterceptor) Option {
	return func(server flux2.WebListener) {
		for _, wi := range array {
			server.AddInterceptor(wi)
		}
	}
}

func WithWebHandlers(tuples []WebHandlerTuple) Option {
	return WithWebHandler(tuples...)
}

func WithWebHandler(tuples ...WebHandlerTuple) Option {
	return func(server flux2.WebListener) {
		for _, h := range tuples {
			server.AddHandler(h.Method, h.Pattern, h.Handler)
		}
	}
}

type WebHandlerTuple struct {
	Method  string
	Pattern string
	Handler flux2.WebHandler
}

func WithHttpHandlers(tuples []HttpHandlerTuple) Option {
	return WithHttpHandler(tuples...)
}

func WithHttpHandler(tuples ...HttpHandlerTuple) Option {
	return func(server flux2.WebListener) {
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
