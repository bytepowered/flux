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

var _ flux.HttpResponseWriter = new(DefaultHttpResponseWriter)

type DefaultHttpResponseWriter struct {
}

func (a *DefaultHttpResponseWriter) WriteError(response *echo.Response,
	requestId string, header http.Header, err *flux.InvokeError) error {
	setupResponse(response, requestId, header, err.StatusCode)
	data := map[string]string{
		"request-id": requestId,
		"status":     "error",
		"message":    err.Message,
	}
	if nil != err.Internal {
		data["error"] = err.Internal.Error()
	}
	return serialize(getHttpSerializer(), response, data)
}

func (a *DefaultHttpResponseWriter) WriteData(
	response *echo.Response,
	requestId string, header http.Header, status int, body interface{}) error {
	resp := setupResponse(response, requestId, header, status)
	if r, ok := body.(io.Reader); ok {
		if c, ok := r.(io.Closer); ok {
			defer pkg.SilentlyCloseFunc(c)
		}
		if data, err := ioutil.ReadAll(r); nil != err {
			return err
		} else {
			return serialize(getHttpSerializer(), resp, data)
		}
	} else {
		return serialize(getHttpSerializer(), resp, body)
	}
}

func serialize(encoder flux.Serializer, resp *echo.Response, data interface{}) error {
	if bytes, err := encoder.Marshal(data); nil != err {
		return &flux.InvokeError{
			StatusCode: flux.StatusServerError,
			Message:    "RESPONSE:MARSHALING",
			Internal:   err,
		}
	} else {
		return writeToHttp(resp, bytes)
	}
}

func getHttpSerializer() flux.Serializer {
	return ext.GetSerializer(ext.TypeNameSerializerDefault)
}

func writeToHttp(resp *echo.Response, bytes []byte) error {
	_, err := resp.Write(bytes)
	if nil != err {
		return fmt.Errorf("write http response: %w", err)
	}
	return err
}

func setupResponse(resp *echo.Response, reqId string, oheader http.Header, status int) *echo.Response {
	resp.Status = status
	headers := resp.Header()
	headers.Set(echo.HeaderServer, "FluxGo")
	headers.Set(echo.HeaderXRequestID, reqId)
	headers.Set("Content-Type", "application/json;charset=utf-8")
	// 允许Override默认Header
	for k, v := range oheader {
		for _, iv := range v {
			headers.Add(k, iv)
		}
	}
	return resp
}
