package listener

import (
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/ext"
	"github.com/bytepowered/flux/flux-node/logger"
	fluxpkg "github.com/bytepowered/flux/flux-pkg"
	"io"
	"io/ioutil"
	"net/http"
)

// DefaultNotfoundHandler 生成NotFound错误，由ErrorHandler处理
func DefaultNotfoundHandler(_ flux.WebExchange) error {
	return &flux.ServeError{
		StatusCode: flux.StatusNotFound,
		ErrorCode:  flux.ErrorCodeRequestNotFound,
		Message:    flux.ErrorMessageWebServerRequestNotFound,
	}
}

func DefaultErrorHandler(webex flux.WebExchange, err error) {
	if nil == err {
		return
	}
	if serr, ok := err.(*flux.ServeError); ok {
		webex.SendError(serr)
	} else {
		webex.SendError(&flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    err.Error(),
			Header:     http.Header{},
			CauseError: err,
		})
	}
}

var _ flux.WebResponseWriter = new(DefaultResponseWriter)

type DefaultResponseWriter struct {
}

func (d *DefaultResponseWriter) Write(webex flux.WebExchange, header http.Header, status int, body interface{}) error {
	fluxpkg.AssertNotNil(body, "<body> is nil, when write body in response writer")
	d.setDefaults(webex, header)
	if bytes, err := Serialize(webex.RequestId(), body); nil != err {
		return webex.Write(status, flux.MIMEApplicationJSON, bytes)
	} else {
		return err
	}
}

func (d *DefaultResponseWriter) WriteError(webex flux.WebExchange, header http.Header, status int, error *flux.ServeError) error {
	fluxpkg.AssertNotNil(error, "<error> is nil, when write error in response writer")
	emap := map[string]interface{}{
		"status":  "error",
		"message": error.Message,
	}
	if nil != error.CauseError {
		emap["error"] = error.CauseError.Error()
	}
	return d.Write(webex, header, status, emap)
}

func (d *DefaultResponseWriter) setDefaults(webex flux.WebExchange, header http.Header) {
	webex.SetResponseHeader(flux.HeaderServer, "Flux/Gateway")
	// 允许Override默认Header
	for k, v := range header {
		for _, iv := range v {
			webex.AddResponseHeader(k, iv)
		}
	}
}

func Serialize(id string, body interface{}) ([]byte, error) {
	if bytes, ok := body.([]byte); ok {
		return bytes, nil
	} else if str, ok := body.(string); ok {
		return []byte(str), nil
	} else if r, ok := body.(io.Reader); ok {
		if c, ok := r.(io.Closer); ok {
			defer func() {
				_ = c.Close()
			}()
		}
		if bytes, err := ioutil.ReadAll(r); nil != err {
			logger.Trace(id).Errorw("SERVER:SERIALIZE:WRITE:reader", "error", err)
			return nil, err
		} else {
			return bytes, nil
		}
	} else {
		if bytes, err := ext.JSONMarshal(body); nil != err {
			logger.Trace(id).Errorw("SERVER:SERIALIZE:WRITE:tojson", "body", body, "error", err)
			return nil, err
		} else {
			return bytes, nil
		}
	}
}
