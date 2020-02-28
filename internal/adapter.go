package internal

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/extension"
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

func (HttpAdapter) pattern(uri string) string {
	// /api/{userId} -> /api/:userId
	replaced := strings.Replace(uri, "}", "", -1)
	if len(replaced) < len(uri) {
		return strings.Replace(replaced, "{", ":", -1)
	} else {
		return uri
	}
}

func (a HttpAdapter) error(c *context, invokeError *flux.InvokeError) error {
	resp := _setupResponse(c, invokeError.StatusCode)
	mdata := map[string]string{
		"requestId": c.RequestId(),
		"status":    "error",
		"message":   invokeError.Message,
	}
	if nil != invokeError.Internal {
		mdata["error"] = invokeError.Internal.Error()
	}
	return _serialize(_formatter(), c, resp, mdata)
}

func (a HttpAdapter) response(c *context) error {
	resp := _setupResponse(c, c.response.status)
	body := c.response.Body()
	if r, ok := body.(io.Reader); ok {
		if c, ok := r.(io.Closer); ok {
			defer pkg.CloseSilently(c)
		}
		if data, err := ioutil.ReadAll(r); nil != err {
			return err
		} else {
			return _serialize(_formatter(), c, resp, data)
		}
	} else {
		return _serialize(_formatter(), c, resp, body)
	}
}

func _serialize(encoder flux.Serializer, c *context, resp *echo.Response, data interface{}) error {
	if bytes, err := encoder.Marshal(data); nil != err {
		return _endErrorTo(encoder, c, resp, &flux.InvokeError{
			StatusCode: flux.StatusServerError,
			Message:    "RESPONSE:MARSHALING",
			Internal:   err,
		})
	} else {
		return _endWriteTo(resp, bytes)
	}
}

func _formatter() flux.Serializer {
	return extension.GetSerializer(extension.TypeNameSerializerDefault)
}

func _endErrorTo(serializer flux.Serializer, c *context, resp *echo.Response, ierr *flux.InvokeError) error {
	data := map[string]string{
		"requestId": c.RequestId(),
		"status":    "error",
		"message":   ierr.Message,
	}
	if nil != ierr.Internal {
		data["error"] = ierr.Internal.Error()
	}
	if bytes, err := serializer.Marshal(data); nil != err {
		return fmt.Errorf("marshal http response: %w", err)
	} else {
		return _endWriteTo(resp, bytes)
	}
}

func _endWriteTo(resp *echo.Response, bytes []byte) error {
	_, err := resp.Write(bytes)
	if nil != err {
		return fmt.Errorf("write http response: %w", err)
	}
	return err
}

func _setupResponse(c *context, status int) *echo.Response {
	resp := c.echo.Response()
	resp.Status = status
	headers := resp.Header()
	headers.Set(echo.HeaderServer, DefaultServerName)
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
