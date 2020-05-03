package flux

import (
	"io"
	"net/http"
)

// RequestReader 定义请求参数读取接口
type RequestReader interface {
	// 获取Http请求的Query参数
	//
	// Deprecated : See QueryValue(name)
	ParamInQuery(name string) string

	// 获取Http请求的Path参数
	//
	// Deprecated : See PathValue(name)
	ParamInPath(name string) string

	// 获取Http请求的Form参数
	//
	// Deprecated : See FormValue(name)
	ParamInForm(name string) string

	// 获取Http请求的Header参数
	//
	// Deprecated : See HeaderValue(name)
	Header(name string) string

	// 获取Http请求的全部Header
	Headers() http.Header

	// 获取Http请求Cookie参数
	//
	// Deprecated : See HeaderValue(name)
	Cookie(name string) string

	// 获取Http请求的Query参数
	QueryValue(name string) string

	// 获取Http请求的Path路径参数
	PathValue(name string) string

	// 获取Http请求的Form表单参数
	FormValue(name string) string

	// 获取Http请求的Header参数
	HeaderValue(name string) string

	// 获取Http请求的Cookie参数
	CookieValue(name string) string

	// 获取Http请求的远程地址
	RemoteAddress() string

	// 返回Http请求的Body可重复读取的接口
	HttpBody() (io.ReadCloser, error)

	// Http原始Request
	HttpRequest() *http.Request
}
