package server

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
	"io"
	"io/ioutil"
	"net/http"
)

var (
	serverWriterSerializer    flux.Serializer
	serverResponseContentType string
)

// SetServerWriterSerializer 设置Http响应数据序列化接口实现；默认为JSON序列化实现。
func SetServerWriterSerializer(s flux.Serializer) {
	serverWriterSerializer = s
}

// GetServerWriterSerializer 获取Http响应数据序列化接口实现；默认为JSON序列化实现。
func GetServerWriterSerializer() flux.Serializer {
	return serverWriterSerializer
}

// SetServerResponseContentType 设置Http响应的MIME类型字符串；默认为JSON/UTF8。
func SetServerResponseContentType(ctype string) {
	serverResponseContentType = ctype
}

// GetServerResponseContentType 获取Http响应的MIME类型字符串；默认为JSON/UTF8。
func GetServerResponseContentType() string {
	return serverResponseContentType
}

func DefaultServerErrorsWriter(webc flux.WebContext, requestId string, header http.Header, serr *flux.ServeError) error {
	SetupResponseDefaults(webc, requestId, header)
	resp := map[string]string{
		"status":  "error",
		"message": serr.Message,
	}
	if nil != serr.Internal {
		resp["error"] = serr.Internal.Error()
	}
	bytes, err := SerializeWith(serverWriterSerializer, resp)
	if nil != err {
		return err
	}
	return WriteHttpResponse(webc, serr.StatusCode, serverResponseContentType, bytes)
}

func DefaultServerResponseWriter(webc flux.WebContext, requestId string, header http.Header, status int, body interface{}) error {
	SetupResponseDefaults(webc, requestId, header)
	var output []byte
	if r, ok := body.(io.Reader); ok {
		if c, ok := r.(io.Closer); ok {
			defer func() {
				_ = c.Close()
			}()
		}
		if bytes, err := ioutil.ReadAll(r); nil != err {
			logger.Trace(requestId).Errorw("Http responseWriter, read body", "error", err)
			return err
		} else {
			output = bytes
		}
	} else {
		if bytes, err := SerializeWith(serverWriterSerializer, body); nil != err {
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
	// 写入Http响应发生的错误，没必要向上抛出Error错误处理。因为已无法通过WriteError写到客户端
	if err := WriteHttpResponse(webc, status, serverResponseContentType, output); nil != err {
		logger.Trace(requestId).Errorw("Http responseWriter, write channel", "data", string(output), "error", err)
	}
	return nil
}

func SerializeWith(serializer flux.Serializer, data interface{}) ([]byte, *flux.ServeError) {
	if bytes, err := serializer.Marshal(data); nil != err {
		return nil, &flux.ServeError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    flux.ErrorMessageWebServerResponseMarshal,
			Internal:   err,
		}
	} else {
		return bytes, nil
	}
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
	webc.SetResponseHeader(flux.HeaderContentType, serverResponseContentType)
	// 允许Override默认Header
	for k, v := range header {
		for _, iv := range v {
			webc.AddResponseHeader(k, iv)
		}
	}
}
