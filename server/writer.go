package server

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/pkg"
	"github.com/bytepowered/flux/webx"
	"io"
	"io/ioutil"
	"net/http"
)

var _ webx.WebServerWriter = new(HttpServerWriter)

// WebServerWriter 默认Http服务响应数据Writer
type HttpServerWriter int

func (a *HttpServerWriter) WriteError(webc webx.WebContext, requestId string, header http.Header, serr *flux.StateError) error {
	SetupResponseDefaults(webc, requestId, header)
	resp := map[string]string{
		"status":  "error",
		"message": serr.Message,
	}
	if nil != serr.Internal {
		resp["error"] = serr.Internal.Error()
	}
	bytes, err := SerializeWith(GetHttpDefaultSerializer(), resp)
	if nil != err {
		return err
	}
	return WriteToHttpChannel(webc, serr.StatusCode, bytes)
}

func (a *HttpServerWriter) WriteBody(webc webx.WebContext, requestId string, header http.Header, status int, body interface{}) error {
	SetupResponseDefaults(webc, requestId, header)
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
	if err := WriteToHttpChannel(webc, status, output); nil != err {
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

func WriteToHttpChannel(webc webx.WebContext, status int, bytes []byte) error {
	_, err := webc.Response().Write(bytes)
	if nil != err {
		return fmt.Errorf("write http response: %w", err)
	}
	webc.Response().WriteHeader(status)
	return err
}

func SetupResponseDefaults(webc webx.WebContext, requestId string, header http.Header) {
	webc.ResponseHeader().Set(webx.HeaderServer, "Flux/Gateway")
	webc.ResponseHeader().Set(flux.XRequestId, requestId)
	webc.ResponseHeader().Set(webx.HeaderContentType, webx.MIMEApplicationJSON)
	// 允许Override默认Header
	for k, v := range header {
		for _, iv := range v {
			webc.ResponseHeader().Add(k, iv)
		}
	}
}
