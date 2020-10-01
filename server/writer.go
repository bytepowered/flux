package server

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/pkg"
	"io"
	"io/ioutil"
	"net/http"
)

func DefaultServerErrorsWriter(webc flux.WebContext, requestId string, header http.Header, serr *flux.StateError) error {
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
	return WriteToHttpChannel(webc, serr.StatusCode, flux.MIMEApplicationJSONCharsetUTF8, bytes)
}

func DefaultServerResponseWriter(webc flux.WebContext, requestId string, header http.Header, status int, body interface{}) error {
	SetupResponseDefaults(webc, requestId, header)
	var output []byte
	if r, ok := body.(io.Reader); ok {
		if c, ok := r.(io.Closer); ok {
			defer pkg.SilentlyCloseFunc(c)
		}
		if bytes, err := ioutil.ReadAll(r); nil != err {
			logger.Trace(requestId).Errorw("Http responseWriter, read body", "error", err)
			return err
		} else {
			output = bytes
		}
	} else {
		if bytes, err := SerializeWith(GetHttpDefaultSerializer(), body); nil != err {
			logger.Trace(requestId).Errorw("Http responseWriter, serialize to json", "body", body, "error", err)
			return err
		} else {
			output = bytes
		}
	}
	// 异步地打印响应日志信息
	go func() {
		logger.Trace(requestId).Infow("Http responseWriter, logging", "data", string(output))
	}()
	// 写入Http响应发生的错误，没必要向上抛出Error错误处理。
	// 因为已无法通过WriteError写到客户端
	if err := WriteToHttpChannel(webc, status, flux.MIMEApplicationJSONCharsetUTF8, output); nil != err {
		logger.Trace(requestId).Errorw("Http responseWriter, write channel", "data", string(output), "error", err)
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

func WriteToHttpChannel(webc flux.WebContext, status int, contentType string, bytes []byte) error {
	err := webc.Write(status, contentType, bytes)
	if nil != err {
		return fmt.Errorf("write http responseWriter: %w", err)
	}
	return err
}

func SetupResponseDefaults(webc flux.WebContext, requestId string, header http.Header) {
	webc.SetResponseHeader(flux.HeaderXRequestId, requestId)
	webc.SetResponseHeader(flux.HeaderServer, "Flux/Gateway")
	webc.SetResponseHeader(flux.HeaderContentType, flux.MIMEApplicationJSONCharsetUTF8)
	// 允许Override默认Header
	for k, v := range header {
		for _, iv := range v {
			webc.AddResponseHeader(k, iv)
		}
	}
}
