package extension

import "github.com/labstack/echo/v4"

var (
	_httpInterceptor = make([]echo.MiddlewareFunc, 0)
	_httpMiddleware  = make([]echo.MiddlewareFunc, 0)
)

// AddHttpInterceptor 添加Http前拦截器。将在Http被路由到对应Handler之前执行
func AddHttpInterceptor(m echo.MiddlewareFunc) {
	_httpInterceptor = append(_httpInterceptor, m)
}

func HttpInterceptors() []echo.MiddlewareFunc {
	s := make([]echo.MiddlewareFunc, len(_httpInterceptor))
	copy(s, _httpInterceptor)
	return s
}

// AddHttpMiddleware 添加Http中间件。在Http路由到对应Handler后执行
func AddHttpMiddleware(m echo.MiddlewareFunc) {
	_httpMiddleware = append(_httpMiddleware, m)
}

func HttpMiddlewares() []echo.MiddlewareFunc {
	s := make([]echo.MiddlewareFunc, len(_httpMiddleware))
	copy(s, _httpMiddleware)
	return s
}
