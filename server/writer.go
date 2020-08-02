package server

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/pkg"
	"github.com/labstack/echo/v4"
	"io"
	"io/ioutil"
	"net/http"
)

var (
	_ flux.HttpResponseWriter = new(HttpServerResponseWriter)
)

// HttpServerResponseWriter 默认Http服务响应数据Writer
type HttpServerResponseWriter int

func (a *HttpServerResponseWriter) WriteError(ctx echo.Context, requestId string, header http.Header, err *flux.StateError) error {
	SetupResponseDefaults(ctx.Response(), requestId, header, err.StatusCode)
	resp := map[string]string{
		"status":  "error",
		"message": err.Message,
	}
	if nil != err.Internal {
		resp["error"] = err.Internal.Error()
	}
	bytes, err := SerializeWith(GetHttpDefaultSerializer(), resp)
	if nil != err {
		return err
	}
	return WriteToHttpChannel(ctx.Response(), bytes)
}

func (a *HttpServerResponseWriter) WriteBody(ctx echo.Context, requestId string, header http.Header, status int, body interface{}) error {
	SetupResponseDefaults(ctx.Response(), requestId, header, status)
	var output []byte
	if r, ok := body.(io.Reader); ok {
		if c, ok := r.(io.Closer); ok {
			defer pkg.SilentlyCloseFunc(c)
		}
		if bytes, err := ioutil.ReadAll(r); nil != err {
			logger.Trace(requestId).Errorw("Http response, read body", "error", err)
			return err
		} else {
			output = bytes
		}
	} else {
		if bytes, err := SerializeWith(GetHttpDefaultSerializer(), body); nil != err {
			logger.Trace(requestId).Errorw("Http response, serialize to json", "body", body, "error", err)
			return err
		} else {
			output = bytes
		}
	}
	// 异步地打印响应日志信息
	go func() {
		logger.Trace(requestId).Infow("Http response, logging", "data", string(output))
	}()
	// 写入Http响应发生的错误，没必要向上抛出Error错误处理。
	// 因为已无法通过WriteError写到客户端
	if err := WriteToHttpChannel(ctx.Response(), output); nil != err {
		logger.Trace(requestId).Errorw("Http response, write channel", "data", string(output), "error", err)
	}
	return nil
}

func SerializeWith(serializer flux.Serializer, data interface{}) ([]byte, *flux.StateError) {
	if bytes, err := serializer.Marshal(data); nil != err {
		return nil, &flux.StateError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    "RESPONSE:MARSHALING",
			Internal:   err,
		}
	} else {
		return bytes, nil
	}
}

func GetHttpDefaultSerializer() flux.Serializer {
	return ext.GetSerializer(ext.TypeNameSerializerDefault)
}

func WriteToHttpChannel(resp *echo.Response, bytes []byte) error {
	_, err := resp.Write(bytes)
	if nil != err {
		return fmt.Errorf("write http response: %w", err)
	}
	return err
}

func SetupResponseDefaults(resp *echo.Response, reqId string, header http.Header, status int) *echo.Response {
	resp.Status = status
	resp.Header().Set(echo.HeaderServer, "Flux/Gateway")
	resp.Header().Set(echo.HeaderXRequestID, reqId)
	resp.Header().Set("Content-Type", "application/json; charset=utf-8")
	// 允许Override默认Header
	for k, v := range header {
		for _, iv := range v {
			resp.Header().Add(k, iv)
		}
	}
	return resp
}
