package listener

import (
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/ext"
	fluxpkg "github.com/bytepowered/flux/flux-pkg"
	"net/http"
)

type (
	Option func(server flux.WebListener)
)

func New(id string, config *flux.Configuration, wis []flux.WebInterceptor, opts ...Option) flux.WebListener {
	opts = append([]Option{
		WithErrorHandler(DefaultErrorHandler),
		WithNotfoundHandler(DefaultNotfoundHandler),
		WithInterceptors(wis),
	}, opts...)
	return NewWith(id, config, opts...)
}

func NewWith(id string, config *flux.Configuration, opts ...Option) flux.WebListener {
	fluxpkg.AssertNotNil(config, "<configuration> of web listener must not nil")
	// 通过Factory创建特定Web框架的Listener
	// 默认为labstack.echo框架：github.com/labstack/echo
	webListener := ext.WebListenerFactory()(id, config)
	for _, opt := range opts {
		opt(webListener)
	}
	return webListener
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
