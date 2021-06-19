package http

import (
	"context"
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/toolkit"
	"github.com/bytepowered/flux/transporter"
	"github.com/spf13/cast"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func DefaultAssembleRequest(ctx *flux.Context, service *flux.Service) (*http.Request, error) {
	// url
	newUrl, newErr := url.Parse(service.Url)
	if newErr != nil {
		return nil, newErr
	}
	// query vars
	newQuery, err := ResolveQueryValues(ctx, service.Arguments)
	if err != nil {
		return nil, fmt.Errorf("resolve query values, error: %w", err)
	}
	if newUrl.RawQuery != "" {
		newUrl.RawQuery += "&" + newQuery.Encode()
	} else {
		newUrl.RawQuery = newQuery.Encode()
	}
	// body
	var newBodyReader io.Reader
	var newContentType = ctx.HeaderVar(flux.HeaderContentType)
	// 如果解析出postforms参数，替换BodyReader
	if postforms, err := ResolvePostFormValues(ctx, service.Arguments); err != nil {
		return nil, fmt.Errorf("resolve form values, error: %w", err)
	} else if len(postforms) > 0 {
		newBodyReader = strings.NewReader(postforms.Encode())
		newContentType = "application/x-www-form-urlencoded"
	} else {
		reader, _ := ctx.BodyReader()
		newBodyReader = reader
	}

	to := service.RpcTimeout()
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
		newRequest.Header.Set("Content-Type", newContentType)
	}
	newRequest.Header.Set("User-Agent", "Flux.go/Transporter/v1")
	return newRequest, err
}

// SelectToArgumentValues 解析Argument参数列表，并返回http标准参数值
func SelectToArgumentValues(ctx *flux.Context, arguments []flux.Argument, selector func(flux.Argument) bool) (url.Values, error) {
	values := make(url.Values, len(arguments))
	for _, arg := range arguments {
		if !selector(arg) {
			continue
		}
		if val, err := transporter.Resolve(ctx, &arg); nil != err {
			return nil, err
		} else {
			values.Add(arg.Name, cast.ToString(val))
		}
	}
	return values, nil
}

// ResolveQueryValues 解析Query参数
func ResolveQueryValues(ctx *flux.Context, args []flux.Argument) (url.Values, error) {
	// 没有定义参数，透传全部Query参数
	if len(args) == 0 {
		return ctx.QueryVars(), nil
	}
	// 已定义，过滤
	return SelectToArgumentValues(ctx, args, func(arg flux.Argument) bool {
		return toolkit.MatchEqual([]string{flux.ScopeQuery, flux.ScopeQueryMulti, flux.ScopeQueryMap}, arg.HttpScope)
	})
}

// ResolvePostFormValues 解析Form表单参数
func ResolvePostFormValues(ctx *flux.Context, args []flux.Argument) (url.Values, error) {
	// 没有定义参数，透传全部Form参数
	if len(args) == 0 {
		return ctx.PostFormVars(), nil
	}
	// 已定义，过滤
	return SelectToArgumentValues(ctx, args, func(arg flux.Argument) bool {
		return toolkit.MatchEqual([]string{flux.ScopeForm, flux.ScopeFormMulti, flux.ScopeFormMap}, arg.HttpScope)
	})
}
