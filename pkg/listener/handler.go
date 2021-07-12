package listener

import (
	ext "github.com/bytepowered/fluxgo/pkg/ext"
	"github.com/bytepowered/fluxgo/pkg/flux"
	"net/http"
)

type (
	Option func(server flux.WebListener)
)

func New(id string, config *flux.Configuration, wis []flux.WebFilter, opts ...Option) flux.WebListener {
	opts = append([]Option{
		WithErrorHandler(DefaultErrorHandler),
		WithNotfoundHandler(DefaultNotfoundHandler),
		WithFilters(wis),
	}, opts...)
	return NewWith(id, config, opts...)
}

func NewWith(id string, config *flux.Configuration, opts ...Option) flux.WebListener {
	flux.AssertNotNil(config, "<configuration> of web listener must not nil")
	// 通过Factory创建特定Web框架的Listener
	// 默认为labstack.echo框架：github.com/labstack/echo
	webListener := ext.WebListenerFactory()(id, config)
	for _, opt := range opts {
		opt(webListener)
	}
	return webListener
}

func WithErrorHandler(handler flux.WebErrorHandlerFunc) Option {
	return func(server flux.WebListener) {
		server.SetErrorHandler(handler)
	}
}

func WithNotfoundHandler(f flux.WebHandlerFunc) Option {
	return func(server flux.WebListener) {
		server.SetNotfoundHandler(f)
	}
}

func WithFilters(array []flux.WebFilter) Option {
	return WithFilter(array...)
}

func WithFilter(array ...flux.WebFilter) Option {
	return func(server flux.WebListener) {
		for _, wi := range array {
			server.AddFilter(wi)
		}
	}
}

func WithHandlers(tuples []WebHandlerTuple) Option {
	return WithHandler(tuples...)
}

func WithHandler(tuples ...WebHandlerTuple) Option {
	return func(server flux.WebListener) {
		for _, h := range tuples {
			server.AddHandler(h.Method, h.Pattern, h.Handler)
		}
	}
}

type WebHandlerTuple struct {
	Method  string
	Pattern string
	Handler flux.WebHandlerFunc
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
