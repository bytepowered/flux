package internal

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/pkg"
	"github.com/labstack/echo/v4"
	"io"
	"io/ioutil"
	"strings"
)

const (
	httpContentTypeJson = "application/json;charset=utf-8"
)

type HttpAdapter int

func (HttpAdapter) Pattern(uri string) string {
	// /api/{userId} -> /api/:userId
	replaced := strings.Replace(uri, "}", "", -1)
	if len(replaced) < len(uri) {
		return strings.Replace(replaced, "{", ":", -1)
	} else {
		return uri
	}
}

func (a HttpAdapter) WriteError(c *Context, invokeError *flux.InvokeError) error {
	resp := _setupResponse(c, invokeError.StatusCode)
	data := map[string]string{
		"requestId": c.RequestId(),
		"status":    "error",
		"message":   invokeError.Message,
	}
	if nil != invokeError.Internal {
		data["error"] = invokeError.Internal.Error()
	}
	return _serializeWith(_httpSerializer(), resp, data)
}

func (a HttpAdapter) WriteResponse(c *Context) error {
	resp := _setupResponse(c, c.response.status)
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

func _setupResponse(c *Context, status int) *echo.Response {
	resp := c.echo.Response()
	resp.Status = status
	headers := resp.Header()
	headers.Set(echo.HeaderServer, "FluxGo")
	headers.Set(echo.HeaderXRequestID, c.RequestId())
	headers.Set("Content-Type", httpContentTypeJson)
	// 允许Override默认Header
	for k, v := range c.response.Headers() {
		for _, iv := range v {
			headers.Add(k, iv)
		}
	}
	return resp
}
