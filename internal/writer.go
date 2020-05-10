package internal

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

type FxHttpWriter int

func (a FxHttpWriter) WriteError(resp *echo.Response, reqId string, outHeader http.Header, invokeError *flux.InvokeError) error {
	_setupResponse(resp, reqId, outHeader, invokeError.StatusCode)
	data := map[string]string{
		"requestId": reqId,
		"status":    "error",
		"message":   invokeError.Message,
	}
	if nil != invokeError.Internal {
		data["error"] = invokeError.Internal.Error()
	}
	return _serializeWith(_httpSerializer(), resp, data)
}

func (a FxHttpWriter) WriteResponse(c *FxContext) error {
	resp := _setupResponse(c.echo.Response(), c.RequestId(), c.ResponseWriter().Headers(), c.response.status)
	body := c.response.Body()
	if r, ok := body.(io.Reader); ok {
		if c, ok := r.(io.Closer); ok {
			defer pkg.CloseSilently(c)
		}
		if data, err := ioutil.ReadAll(r); nil != err {
			return err
		} else {
			return _serializeWith(_httpSerializer(), resp, data)
		}
	} else {
		return _serializeWith(_httpSerializer(), resp, body)
	}
}

func _serializeWith(encoder flux.Serializer, resp *echo.Response, data interface{}) error {
	if bytes, err := encoder.Marshal(data); nil != err {
		return &flux.InvokeError{
			StatusCode: flux.StatusServerError,
			Message:    "RESPONSE:MARSHALING",
			Internal:   err,
		}
	} else {
		return _writeToHttp(resp, bytes)
	}
}

func _httpSerializer() flux.Serializer {
	return ext.GetSerializer(ext.TypeNameSerializerDefault)
}

func _writeToHttp(resp *echo.Response, bytes []byte) error {
	_, err := resp.Write(bytes)
	if nil != err {
		return fmt.Errorf("write http response: %w", err)
	}
	return err
}

func _setupResponse(resp *echo.Response, reqId string, outHeader http.Header, status int) *echo.Response {
	resp.Status = status
	headers := resp.Header()
	headers.Set(echo.HeaderServer, "FluxGo")
	headers.Set(echo.HeaderXRequestID, reqId)
	headers.Set("Content-Type", "application/json;charset=utf-8")
	// 允许Override默认Header
	for k, v := range outHeader {
		for _, iv := range v {
			headers.Add(k, iv)
		}
	}
	return resp
}
