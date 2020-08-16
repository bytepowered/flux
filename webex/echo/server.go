package echo

import (
	"context"
	"github.com/bytepowered/flux/webex"
	"github.com/labstack/echo/v4"
)

var _ webex.WebServer = new(ServerEchoAdapter)

func NewServerEchoAdapter() webex.WebServer {
	return &ServerEchoAdapter{
		server: echo.New(),
	}
}

type ServerEchoAdapter struct {
	server *echo.Echo
}

func (w *ServerEchoAdapter) SetNotFoundHandler(fun webex.HandlerFunc) {
	echo.NotFoundHandler = func(c echo.Context) error {
		return fun(NewAdaptEchoContext(c))
	}
}

func (w *ServerEchoAdapter) SetErrorHandler(fun webex.ErrorHandlerFunc) {
	w.server.HTTPErrorHandler = func(err error, c echo.Context) {
		fun(err, NewAdaptEchoContext(c))
	}
}

func (w *ServerEchoAdapter) AddInterceptor(m webex.MiddlewareFunc) {
	w.server.Pre(ToAdaptMiddlewareFunc(m))
}

func (w *ServerEchoAdapter) AddMiddleware(m webex.MiddlewareFunc) {
	w.server.Use(ToAdaptMiddlewareFunc(m))
}

func (w *ServerEchoAdapter) AddRouteHandler(method, pattern string, h webex.HandlerFunc, m ...webex.MiddlewareFunc) {
	ms := make([]echo.MiddlewareFunc, len(m))
	for i, mi := range m {
		ms[i] = ToAdaptMiddlewareFunc(mi)
	}
	w.server.Add(method, pattern, ToAdaptHandlerFunc(h), ms...)
}

func (w *ServerEchoAdapter) Start(addr string) error {
	return w.server.Start(addr)
}

func (w *ServerEchoAdapter) StartTLS(addr string, certFile, keyFile string) error {
	return w.server.StartTLS(addr, certFile, keyFile)
}

func (w *ServerEchoAdapter) Shutdown(ctx context.Context) error {
	return w.server.Shutdown(ctx)
}
