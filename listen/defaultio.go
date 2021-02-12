package listen

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"io"
	"io/ioutil"
	"net/http"
)

// DefaultNotfoundHandler 生成NotFound错误，由ErrorHandler处理
func DefaultNotfoundHandler(_ flux.WebContext) error {
	return &flux.ServeError{
		StatusCode: flux.StatusNotFound,
		ErrorCode:  flux.ErrorCodeRequestNotFound,
		Message:    flux.ErrorMessageWebServerRequestNotFound,
	}
}

func DefaultServerErrorHandler(webc flux.WebContext, err error) {
	if nil == err {
		return
	}
	if serr, ok := err.(*flux.ServeError); ok {
		webc.SendError(serr)
	} else {
		webc.SendError(&flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    err.Error(),
			Header:     http.Header{},
			Internal:   err,
		})
	}
}

func DefaultResponseWriter(webc flux.WebContext, header http.Header, status int, body interface{}, serr *flux.ServeError) error {
	id := webc.ScopeValue(flux.HeaderXRequestId).(string)
	SetupResponseDefaults(webc, id, header)
	var payload interface{}
	if nil != serr {
		emap := map[string]interface{}{
			"status":  "error",
			"message": serr.Message,
		}
		if nil != serr.Internal {
			emap["error"] = serr.Internal.Error()
		}
		payload = emap
	} else {
		payload = body
	}
	// 序列化payload
	var data []byte
	if r, ok := payload.(io.Reader); ok {
		if c, ok := r.(io.Closer); ok {
			defer func() {
				_ = c.Close()
			}()
		}
		if bytes, err := ioutil.ReadAll(r); nil != err {
			logger.With(id).Errorw("Http-ResponseWriter, read body", "error", err)
			return err
		} else {
			data = bytes
		}
	} else {
		if bytes, err := ext.JSONMarshal(payload); nil != err {
			logger.With(id).Errorw("Http-ResponseWriter, serialize to json", "body", payload, "error", err)
			return err
		} else {
			data = bytes
		}
	}
	logger.With(id).Infow("Http-ResponseWriter, logging", "data", string(data))
	// 写入Http响应发生的错误，没必要向上抛出Error错误处理。因为已无法通过WriteError写到客户端
	if err := WriteHttpResponse(webc, status, flux.MIMEApplicationJSON, data); nil != err {
		logger.With(id).Errorw("Http-ResponseWriter, write channel", "data", string(data), "error", err)
	}
	return nil
}

func WriteHttpResponse(webc flux.WebContext, statusCode int, contentType string, bytes []byte) error {
	err := webc.Write(statusCode, contentType, bytes)
	if nil != err {
		return fmt.Errorf("write http responseWriter: %w", err)
	}
	return err
}

func SetupResponseDefaults(webc flux.WebContext, requestId string, header http.Header) {
	webc.SetResponseHeader(flux.HeaderXRequestId, requestId)
	webc.SetResponseHeader(flux.HeaderServer, "Flux/Gateway")
	webc.SetResponseHeader(flux.HeaderContentType, flux.MIMEApplicationJSON)
	// 允许Override默认Header
	for k, v := range header {
		for _, iv := range v {
			webc.AddResponseHeader(k, iv)
		}
	}
}
