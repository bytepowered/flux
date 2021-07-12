package listener

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
)

import (
	"github.com/bytepowered/fluxgo/pkg/ext"
	"github.com/bytepowered/fluxgo/pkg/flux"
	"github.com/bytepowered/fluxgo/pkg/logger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/random"
)

// DefaultRequestBodyResolver 默认对RequestBody的表单数据进行解析，返回Post表单列表
func DefaultRequestBodyResolver(webex flux.WebContext) url.Values {
	return webex.PostFormVars()
}

// DefaultRequestIdentifierLocator 默认RequestId查找函数实现
func DefaultRequestIdentifierLocator(ctx interface{}) string {
	echoc, ok := ctx.(echo.Context)
	flux.Assert(ok, "<context> must be echo.context")
	id := echoc.Request().Header.Get(flux.XRequestId)
	if "" != id {
		return id
	}
	echoc.Request().Header.Set("X-RequestId-By", "flux")
	return "fxid_" + random.String(32)
}

// RepeatableReadFilter Body缓存，允许通过 GetBody 多次读取Body
func RepeatableReadFilter(next echo.HandlerFunc) echo.HandlerFunc {
	// 包装Http处理错误，统一由HttpErrorHandler处理
	return func(echo echo.Context) error {
		request := echo.Request()
		data, err := ioutil.ReadAll(request.Body)
		if nil != err {
			return &flux.ServeError{
				StatusCode: flux.StatusBadRequest,
				ErrorCode:  flux.ErrorCodeGatewayInternal,
				Message:    flux.ErrorMessageRequestPrepare,
				CauseError: fmt.Errorf("read request body, method: %s, uri:%s, err: %w", request.Method, request.RequestURI, err),
			}
		}
		request.GetBody = func() (io.ReadCloser, error) {
			return ioutil.NopCloser(bytes.NewBuffer(data)), nil
		}
		// 恢复Body，但ParseForm解析后，request.Body无法重读，需要通过GetBody
		request.Body = ioutil.NopCloser(bytes.NewBuffer(data))
		return next(echo)
	}
}

// DefaultNotfoundHandler 生成NotFound错误，由ErrorHandler处理
func DefaultNotfoundHandler(_ flux.WebContext) error {
	return &flux.ServeError{
		StatusCode: flux.StatusNotFound,
		ErrorCode:  flux.ErrorCodeRequestNotFound,
		Message:    flux.ErrorMessageWebServerRequestNotFound,
	}
}

func DefaultErrorHandler(webex flux.WebContext, error error) {
	if flux.IsNil(error) {
		return
	}
	serr, ok := error.(*flux.ServeError)
	if !ok {
		serr = &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    error.Error(),
			CauseError: error,
		}
	}
	data, err := ext.JSONMarshalObject(serr)
	if err != nil {
		logger.Trace(webex.RequestId()).Errorw("SERVER:ERROR_HANDLE", "error", err)
		return
	}
	webex.ResponseWriter().Header().Add("X-Writer-Id", "Fx-EWriter")
	if err := webex.Write(serr.StatusCode, flux.MIMEApplicationJSON, data); nil != err {
		logger.Trace(webex.RequestId()).Errorw("SERVER:ERROR_HANDLE", "error", err)
	}
}
