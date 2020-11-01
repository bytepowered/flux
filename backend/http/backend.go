package http

import (
	"context"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/support"
	"github.com/spf13/cast"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func NewHttpBackend() *HttpBackend {
	return &HttpBackend{
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

type HttpBackend struct {
	httpClient *http.Client
}

func (ex *HttpBackend) Exchange(ctx flux.Context) *flux.StateError {
	return support.InvokeBackendExchange(ctx, ex)
}

func (ex *HttpBackend) Invoke(service flux.Service, ctx flux.Context) (interface{}, *flux.StateError) {
	inURL, _ := ctx.Request().RequestURL()
	bodyReader, _ := ctx.Request().RequestBodyReader()
	newRequest, err := ex.Assemble(&service, inURL, bodyReader, ctx.Context())
	if nil != err {
		return nil, &flux.StateError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    "HTTPEX:ASSEMBLE",
			Internal:   err,
		}
	} else {
		// Header透传以及传递AttrValues
		if header, writable := ctx.Request().HeaderValues(); writable {
			newRequest.Header = header.Clone()
		} else {
			newRequest.Header = header
		}
		for k, v := range ctx.Attributes() {
			newRequest.Header.Set(k, cast.ToString(v))
		}
	}
	resp, err := ex.httpClient.Do(newRequest)
	if nil != err {
		msg := "HTTPEX:REMOTE_ERROR"
		if uErr, ok := err.(*url.Error); ok {
			msg = fmt.Sprintf("HTTPEX:REMOTE_ERROR:%s", uErr.Error())
		}
		return nil, &flux.StateError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayBackend,
			Message:    msg,
			Internal:   err,
		}
	}
	return resp, nil
}

func (ex *HttpBackend) Assemble(service *flux.Service, inURL *url.URL, bodyReader io.ReadCloser, ctx context.Context) (*http.Request, error) {
	inParams := service.Arguments
	newQuery := inURL.RawQuery
	// 使用可重复读的GetBody函数
	defer func() {
		_ = bodyReader.Close()
	}()
	var newBodyReader io.Reader = bodyReader
	if len(inParams) > 0 {
		// 如果Endpoint定义了参数，即表示限定参数传递
		data := _toHttpUrlValues(inParams).Encode()
		// GET：参数拼接到URL中；
		if http.MethodGet == service.Method {
			if newQuery == "" {
				newQuery = data
			} else {
				newQuery += "&" + data
			}
		} else {
			// 其它方法：拼接到Body中，并设置form-data/x-www-url-encoded
			newBodyReader = strings.NewReader(data)
		}
	}
	// 未定义参数，即透传Http请求：Rewrite inRequest path
	newUrl := &url.URL{
		Host:       service.Host,
		Path:       service.Interface,
		Scheme:     inURL.Scheme,
		Opaque:     inURL.Opaque,
		User:       inURL.User,
		RawPath:    inURL.RawPath,
		ForceQuery: inURL.ForceQuery,
		RawQuery:   newQuery,
		Fragment:   inURL.Fragment,
	}
	timeout, err := time.ParseDuration(service.Timeout)
	if err != nil {
		logger.Warnf("Illegal endpoint rpc-timeout: ", service.Timeout)
		timeout = time.Second * 10
	}
	toctx, _ := context.WithTimeout(ctx, timeout)
	newRequest, err := http.NewRequestWithContext(toctx, service.Method, newUrl.String(), newBodyReader)
	if nil != err {
		return nil, fmt.Errorf("new request, method: %s, url: %s, err: %w", service.Method, newUrl, err)
	}
	// Body数据设置application/x-www-url-encoded
	if http.MethodGet != service.Method {
		newRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	newRequest.Header.Set("User-Agent", "FluxGo/Backend/v1")
	return newRequest, err
}
