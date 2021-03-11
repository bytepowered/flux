package http

import (
	"context"
	"fmt"
	flux2 "github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/logger"
	"github.com/spf13/cast"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func DefaultArgumentAssemble(service *flux2.BackendService, inURL *url.URL, bodyReader io.ReadCloser, ctx flux2.Context) (*http.Request, error) {
	inParams := service.Arguments
	newQuery := inURL.RawQuery
	// 使用可重复读的GetBody函数
	defer func() {
		_ = bodyReader.Close()
	}()
	var newBodyReader io.Reader = bodyReader
	if len(inParams) > 0 {
		// 如果Endpoint定义了参数，即表示限定参数传递
		var data string
		if values, err := AssembleHttpValues(inParams, ctx); nil != err {
			return nil, err
		} else {
			data = values.Encode()
		}
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
		Host:       service.RemoteHost,
		Path:       service.Interface,
		Scheme:     service.Scheme,
		Opaque:     inURL.Opaque,
		User:       inURL.User,
		RawPath:    inURL.RawPath,
		ForceQuery: inURL.ForceQuery,
		RawQuery:   newQuery,
		Fragment:   inURL.Fragment,
	}
	to := service.AttrRpcTimeout()
	timeout, err := time.ParseDuration(to)
	if err != nil {
		logger.Warnf("Illegal endpoint rpc-timeout: ", to)
		timeout = time.Second * 10
	}
	toctx, _ := context.WithTimeout(ctx.Context(), timeout)
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

func AssembleHttpValues(arguments []flux2.Argument, ctx flux2.Context) (url.Values, error) {
	values := make(url.Values, len(arguments))
	for _, arg := range arguments {
		if val, err := arg.Resolve(ctx); nil != err {
			return nil, err
		} else {
			values.Add(arg.Name, cast.ToString(val))
		}
	}
	return values, nil
}
