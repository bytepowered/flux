package filter

import (
	"errors"
	"fmt"
	"github.com/bytepowered/cache"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/spf13/cast"
	"reflect"
	"strings"
	"time"
)

const (
	TypeIdEndpointPermission = "EndpointPermissionFilter"
)

func init() {
	SetEndpointPermissionResponseDecoder(defaultEndpointPermissionDecoder)
}

var (
	endpointPermissionResponseDecoder EndpointPermissionResponseDecoder
)

func SetEndpointPermissionResponseDecoder(decoder EndpointPermissionResponseDecoder) {
	endpointPermissionResponseDecoder = decoder
}

func GetEndpointPermissionResponseDecoder() EndpointPermissionResponseDecoder {
	return endpointPermissionResponseDecoder
}

// 权限验证结果解析函数
type EndpointPermissionResponseDecoder func(response interface{}, ctx flux.Context) (pass bool, expire time.Duration, err error)

func EndpointPermissionFactory() interface{} {
	return &EndpointPermissionFilter{}
}

// EndpointPermissionFilter 提供基于EndpointPermission的权限验证
type EndpointPermissionFilter struct {
	disabled        bool
	permissionCache cache.Cache
}

func (p *EndpointPermissionFilter) DoFilter(next flux.FilterHandler) flux.FilterHandler {
	if p.disabled {
		return next
	}
	return func(ctx flux.Context) *flux.StateError {
		// 必须开启Authorize才进行权限校验
		endpoint := ctx.Endpoint()
		permission := endpoint.Permission
		if false == endpoint.Authorize || !permission.IsValid() {
			return next(ctx)
		}
		// Lookup & resolve arguments
		lookup := ext.GetArgumentValueLookupFunc()
		resolver := ext.GetArgumentValueResolveFunc()
		argsKey, err := _newArgumentsKey(permission.Arguments, lookup, resolver, ctx)
		if nil != err {
			return &flux.StateError{
				StatusCode: flux.StatusServerError,
				Message:    "PERMISSION:LOOKUP/RESOLVE:ERROR",
				Internal:   err,
			}
		}
		// 以Permission的(UpstreamServiceTag + 具体参数Value列表)来构建单个请求的缓存Key
		serviceTag := flux.NewServiceKey(permission.RpcProto, permission.RemoteHost, permission.Method, permission.Interface)
		cacheKey := serviceTag + "#" + argsKey
		// 权限验证结果缓存
		passed, err := p.permissionCache.GetOrLoad(cacheKey, func(_ interface{}) (interface{}, *time.Duration, error) {
			return p.doPermissionVerification(&permission, ctx)
		})
		if nil != err {
			if serr, ok := err.(*flux.StateError); ok {
				return serr
			} else {
				return &flux.StateError{
					StatusCode: flux.StatusServerError,
					Message:    "PERMISSION:LOAD:ERROR",
					Internal:   err,
				}
			}
		}
		if !cast.ToBool(passed) {
			return err.(*flux.StateError)
		} else {
			return next(ctx)
		}
	}
}

func (p *EndpointPermissionFilter) Init(config *flux.Configuration) error {
	config.SetDefaults(map[string]interface{}{
		ConfigKeyCacheExpiration: defValueCacheExpiration,
		ConfigKeyCacheSize:       defValueCacheSize,
		ConfigKeyDisabled:        false,
	})
	p.disabled = config.GetBool(ConfigKeyDisabled)
	if p.disabled {
		logger.Info("Endpoint permission filter was DISABLED!!")
		return nil
	}
	expiration := config.GetDuration(ConfigKeyCacheExpiration)
	cacheSize := config.GetInt(ConfigKeyCacheSize)
	p.permissionCache = cache.New(cacheSize).Expiration(expiration).LRU().Build()
	logger.Infow("Endpoint permission filter init", "cache-alg", "ExpireLRU", "cache-size", cacheSize, "cache-expire", expiration.String())
	return nil
}

func (p *EndpointPermissionFilter) doPermissionVerification(perm *flux.PermissionService, ctx flux.Context) (pass bool, expire *time.Duration, err *flux.StateError) {
	backend, ok := ext.GetBackend(perm.RpcProto)
	if !ok {
		logger.TraceContext(ctx).Errorw("Provider backend unsupported protocol",
			"provider-proto", perm.RpcProto, "provider-uri", perm.Interface, "provider-method", perm.Method)
		return false, cache.NoExpiration, &flux.StateError{
			StatusCode: flux.StatusServerError,
			Message:    "PERMISSION:PROVIDER:UNKNOWN_PROTOCOL",
			Internal:   err,
		}
	}
	service := flux.BackendService{
		RemoteHost: perm.RemoteHost,
		Method:     perm.Method,
		Interface:  perm.Interface,
		Arguments:  perm.Arguments,
	}
	if ret, err := backend.Invoke(service, ctx); nil != err {
		logger.TraceContext(ctx).Errorw("Permission Provider backend load error",
			"provider-proto", perm.RpcProto, "provider-uri", perm.Interface, "provider-method", perm.Method, "error", err)
		return false, cache.NoExpiration, &flux.StateError{
			StatusCode: flux.StatusServerError,
			Message:    "PERMISSION:PROVIDER:LOAD",
			Internal:   err,
		}
	} else {
		passed, expire, err := GetEndpointPermissionResponseDecoder()(ret, ctx)
		if nil != err {
			logger.TraceContext(ctx).Errorw("Permission decode response error",
				"provider-proto", perm.RpcProto, "provider-uri", perm.Interface, "provider-method", perm.Method, "error", err)
			return false, cache.NoExpiration, &flux.StateError{
				StatusCode: flux.StatusServerError,
				Message:    "PERMISSION:RESPONSE:DECODE",
				Internal:   err,
			}
		} else {
			return passed, &expire, nil
		}
	}
}

func (p *EndpointPermissionFilter) Order() int {
	return OrderFilterEndpointPermission
}

func (*EndpointPermissionFilter) TypeId() string {
	return TypeIdEndpointPermission
}

func defaultEndpointPermissionDecoder(response interface{}, ctx flux.Context) (bool, time.Duration, error) {
	logger.TraceContext(ctx).Infow("Decode endpoint permission",
		"response-type", reflect.TypeOf(response), "response", response)
	// 默认支持响应JSON数据：
	// {"status": "[success,error]", "permission": "[true,false]", "message": "OnErrorMessage", "expire": 5}
	strmap, ok := response.(map[string]interface{})
	if ok {
		if "success" == cast.ToString(strmap["status"]) {
			passed := cast.ToBool(strmap["permission"])
			minutes := cast.ToInt(strmap["expire"])
			if minutes < 1 {
				minutes = 1
			}
			return passed, time.Minute * time.Duration(minutes), nil
		} else {
			message := cast.ToString(strmap["message"])
			if "" == message {
				message = "Permission NOT SUCCESS, error message NOT FOUND"
			}
			return false, time.Duration(0), errors.New(message)
		}
	}
	// 如果不是默认JSON结构的数据，只是包含success字符串，就是验证成功
	text := cast.ToString(response)
	pass := strings.Contains(text, "success")
	return pass, time.Minute * 5, nil
}

func _newArgumentsKey(args []flux.Argument,
	lookup flux.ArgumentValueLookupFunc, resolver flux.ArgumentValueResolveFunc, ctx flux.Context) (string, error) {
	// [(T:v1),(T:v2),]
	sb := new(strings.Builder)
	sb.WriteByte('[')
	for _, arg := range args {
		if sv, err := _newArgumentKey(arg, lookup, resolver, ctx); nil != err {
			return "", err
		} else {
			sb.WriteString(sv)
		}
		sb.WriteByte(',')
	}
	sb.WriteByte(']')
	return sb.String(), nil
}

func _newArgumentKey(arg flux.Argument, lookup flux.ArgumentValueLookupFunc, resolver flux.ArgumentValueResolveFunc, ctx flux.Context) (string, error) {
	// (T:val)
	sb := new(strings.Builder)
	sb.WriteByte('(')
	sb.WriteString(arg.Class)
	sb.WriteByte(':')
	if flux.ArgumentTypeComplex == arg.Type && len(arg.Fields) > 0 {
		if sv, err := _newArgumentsKey(arg.Fields, lookup, resolver, ctx); nil != err {
			return "", err
		} else {
			sb.WriteString(sv)
		}
	} else {
		mtValue, err := lookup(arg.HttpScope, arg.HttpName, ctx)
		if nil != err {
			logger.TraceContext(ctx).Warnw("Failed to lookup argument",
				"http.key", arg.HttpName, "arg.name", arg.Name, "error", err)
			return "", fmt.Errorf("ARGUMENT:LOOKUP:%w", err)
		}
		if sv, err := resolver(mtValue, arg, ctx); nil != err {
			return "", err
		} else {
			sb.WriteString(cast.ToString(sv))
		}
	}
	sb.WriteByte(')')
	return sb.String(), nil
}
