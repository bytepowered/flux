package webfast

import (
	"context"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"net/http"
)

var _ flux.WebServer = new(AdaptWebServer)

func init() {
	ext.SetWebServerFactory(NewAdaptWebServer)
}

func NewAdaptWebServer() flux.WebServer {
	// TODO webmidware
	panic("implement me")
	r := router.New()
	return &AdaptWebServer{server: &fasthttp.Server{
		Handler: r.Handler,
	}}
}

type AdaptWebServer struct {
	router *router.Router
	server *fasthttp.Server
}

func (w *AdaptWebServer) SetWebErrorHandler(fun flux.WebErrorHandler) {
	w.server.ErrorHandler = func(ctx *fasthttp.RequestCtx, err error) {
		fun(err, toAdaptWebContext(ctx))
	}
	w.router.PanicHandler = func(ctx *fasthttp.RequestCtx, re interface{}) {
		webc := toAdaptWebContext(ctx)
		if err, ok := re.(error); ok {
			fun(err, webc)
		} else {
			fun(fmt.Errorf("panic: %v", re), webc)
		}
	}
}

func (w *AdaptWebServer) SetRouteNotFoundHandler(fun flux.WebRouteHandler) {
	w.router.NotFound = func(ctx *fasthttp.RequestCtx) {
		if err := fun(toAdaptWebContext(ctx)); nil != err {
			w.server.ErrorHandler(ctx, err)
		}
	}
}

func (w *AdaptWebServer) AddWebInterceptor(m flux.WebMiddleware) {
	// TODO interceptor
	panic("implement me")

}

func (w *AdaptWebServer) AddWebMiddleware(m flux.WebMiddleware) {
	// TODO webmidware
	panic("implement me")
}

func (w *AdaptWebServer) AddStdHttpHandler(method, pattern string, h http.Handler, m ...func(http.Handler) http.Handler) {
	//w.router.Handle(method, pattern, fasthttpadaptor.NewFastHTTPHandler(h))
	// TODO
	panic("implement me")
}

func (w *AdaptWebServer) AddWebRouteHandler(method, pattern string, fun flux.WebRouteHandler, m ...flux.WebMiddleware) {
	h := func(ctx *fasthttp.RequestCtx) {
		webc := toAdaptWebContext(ctx)
		if err := fun(webc); nil != err {
			w.server.ErrorHandler(ctx, err)
		}
	}
	w.router.Handle(method, pattern, h)
	// TODO webmidware
	panic("implement me")
}

func (w *AdaptWebServer) WebRouter() interface{} {
	return w.router
}

func (w *AdaptWebServer) WebServer() interface{} {
	return w.server
}

func (w *AdaptWebServer) Start(addr string) error {
	return w.server.ListenAndServe(addr)
}

func (w *AdaptWebServer) StartTLS(addr string, certFile, keyFile string) error {
	return w.server.ListenAndServeTLS(addr, certFile, keyFile)
}

func (w *AdaptWebServer) Shutdown(_ context.Context) error {
	return w.server.Shutdown()
}
