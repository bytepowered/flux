package server

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/pkg"
	"github.com/labstack/echo/v4"
	"io"
	"io/ioutil"
	"net/http"
)

var (
	_                     flux.HttpResponseWriter = new(HttpServerResponseWriter)
	httpErrorAssembleFunc InvokeErrorAssembleFunc
)

func init() {
	SetHttpErrorAssembleFunc(func(err *flux.InvokeError) map[string]string {
		ssmap := map[string]string{
			"status":  "error",
			"message": err.Message,
		}
		if nil != err.Internal {
			ssmap["error"] = err.Internal.Error()
		}
		return ssmap
	})
}

// InvokeErrorAssembleFunc 将Error转换成响应结构体
type InvokeErrorAssembleFunc func(*flux.InvokeError) map[string]string

// SetHttpErrorAssembleFunc 设置HttpError错误组装成Map响应结构体的处理函数
func SetHttpErrorAssembleFunc(f InvokeErrorAssembleFunc) {
	httpErrorAssembleFunc = f
}

// HttpServerResponseWriter 默认Http服务响应数据Writer
type HttpServerResponseWriter int

func (a *HttpServerResponseWriter) WriteError(response *echo.Response, requestId string, header http.Header, err *flux.InvokeError) error {
	SetupResponseDefaults(response, requestId, header, err.StatusCode)
	bytes, err := SerializeWith(GetHttpDefaultSerializer(), httpErrorAssembleFunc(err))
	if nil != err {
		return err
	}
	return WriteToHttpChannel(response, bytes)
}

func (a *HttpServerResponseWriter) WriteBody(response *echo.Response, requestId string, header http.Header, status int, body interface{}) error {
	SetupResponseDefaults(response, requestId, header, status)
	if r, ok := body.(io.Reader); ok {
		if c, ok := r.(io.Closer); ok {
			defer pkg.SilentlyCloseFunc(c)
		}
		if data, err := ioutil.ReadAll(r); nil != err {
			return err
		} else {
			bytes, err := SerializeWith(GetHttpDefaultSerializer(), data)
			if nil != err {
				return err
			}
			return WriteToHttpChannel(response, bytes)
		}
	} else {
		bytes, err := SerializeWith(GetHttpDefaultSerializer(), body)
		if nil != err {
			return err
		}
		return WriteToHttpChannel(response, bytes)
	}
}

func SerializeWith(serializer flux.Serializer, data interface{}) ([]byte, *flux.InvokeError) {
	if bytes, err := serializer.Marshal(data); nil != err {
		return nil, &flux.InvokeError{
			StatusCode: flux.StatusServerError,
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
	resp.Header().Set(echo.HeaderServer, "FluxGateway")
	resp.Header().Set(echo.HeaderXRequestID, reqId)
	resp.Header().Set("Content-Type", "application/json;charset=utf-8")
	// 允许Override默认Header
	for k, v := range header {
		for _, iv := range v {
			resp.Header().Add(k, iv)
		}
	}
	return resp
}