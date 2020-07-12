package server

import (
	"bytes"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/labstack/echo/v4"
	"io"
	"io/ioutil"
)

// Body缓存，允许通过 GetBody 多次读取Body
func RepeatableBody(next echo.HandlerFunc) echo.HandlerFunc {
	// 包装Http处理错误，统一由HttpErrorHandler处理
	return func(c echo.Context) error {
		request := c.Request()
		data, err := ioutil.ReadAll(request.Body)
		if nil != err {
			return &flux.InvokeError{
				StatusCode: flux.StatusBadRequest,
				ErrorCode:  flux.ErrorCodeGatewayInternal,
				Message:    "REQUEST:BODY_PREPARE",
				Internal:   fmt.Errorf("read req-body, method: %s, uri:%s, err: %w", request.Method, request.RequestURI, err),
			}
		}
		request.GetBody = func() (io.ReadCloser, error) {
			return ioutil.NopCloser(bytes.NewBuffer(data)), nil
		}
		// 恢复Body，但ParseForm解析后，request.Body无法重读，需要通过GetBody
		request.Body = ioutil.NopCloser(bytes.NewBuffer(data))
		if err := request.ParseForm(); nil != err {
			return &flux.InvokeError{
				StatusCode: flux.StatusBadRequest,
				ErrorCode:  flux.ErrorCodeGatewayInternal,
				Message:    "REQUEST:FORM_PARSING",
				Internal:   fmt.Errorf("parsing req-form, method: %s, uri:%s, err: %w", request.Method, request.RequestURI, err),
			}
		} else {
			return next(c)
		}
	}
}
