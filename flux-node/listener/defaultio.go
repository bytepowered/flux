package listener

import (
	"fmt"
	"github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/ext"
	"github.com/bytepowered/flux/flux-node/logger"
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

func DefaultResponseWriter(webex flux.WebExchange, header http.Header, status int, body interface{}, serr *flux.ServeError) error {
	SetupResponseDefaults(webex, header)
	var payload interface{}
	if nil != serr {
		emap := map[string]interface{}{
			"status":  "error",
			"message": serr.Message,
		}
		if nil != serr.CauseError {
			emap["error"] = serr.CauseError.Error()
		}
		payload = emap
	} else {
		payload = body
	}
	id := webex.RequestId()
	// 序列化payload
	var data []byte
	if r, ok := payload.(io.Reader); ok {
		if c, ok := r.(io.Closer); ok {
			defer func() {
				_ = c.Close()
			}()
		}
		if bytes, err := ioutil.ReadAll(r); nil != err {
			logger.Trace(id).Errorw("Http-ResponseWriter, read body", "error", err)
			return err
		} else {
			data = bytes
		}
	} else {
		if bytes, err := ext.JSONMarshal(payload); nil != err {
			logger.Trace(id).Errorw("Http-ResponseWriter, serialize to json", "body", payload, "error", err)
			return err
		} else {
			data = bytes
		}
	}
	logger.Trace(id).Infow("Http-ResponseWriter, logging", "data", string(data))
	// 写入Http响应发生的错误，没必要向上抛出Error错误处理。因为已无法通过WriteError写到客户端
	if err := WriteHttpResponse(webex, status, flux.MIMEApplicationJSON, data); nil != err {
		logger.Trace(id).Errorw("Http-ResponseWriter, write channel", "data", string(data), "error", err)
	}
	return nil
}

func WriteHttpResponse(webex flux.WebExchange, statusCode int, contentType string, bytes []byte) error {
	err := webex.Write(statusCode, contentType, bytes)
	if nil != err {
		return fmt.Errorf("write http responseWriter: %w", err)
	}
	return err
}

func SetupResponseDefaults(webex flux.WebExchange, header http.Header) {
	webex.SetResponseHeader(flux.HeaderServer, "Flux/Gateway")
	webex.SetResponseHeader(flux.HeaderContentType, flux.MIMEApplicationJSON)
	// 允许Override默认Header
	for k, v := range header {
		for _, iv := range v {
			webex.AddResponseHeader(k, iv)
		}
	}
}
