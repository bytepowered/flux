package filter

import (
	"errors"
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
		// 以Permission的(UpstreamServiceTag + 具体参数Value列表)来构建单个请求的缓存Key
		serviceTag := flux.NewServiceKey(permission.UpstreamProto, permission.UpstreamHost, permission.UpstreamMethod, permission.UpstreamUri)
		cacheKey := serviceTag + "#" + _newArgumentsKey(permission.Arguments)
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
	logger.Infow("Endpoint permission filter init", "cache-alg", "ExpireLRU", "cache-size", cacheSize, "cache-expire", expiration)
	return nil
}

func (p *EndpointPermissionFilter) doPermissionVerification(meta *flux.Permission, ctx flux.Context) (pass bool, expire *time.Duration, err *flux.StateError) {
	provider, ok := ext.GetBackend(meta.UpstreamProto)
	if !ok {
		logger.Trace(ctx.RequestId()).Errorw("Provider backend unsupported protocol",
			"provider-proto", meta.UpstreamProto, "provider-uri", meta.UpstreamUri, "provider-method", meta.UpstreamMethod)
		return false, cache.NoExpiration, &flux.StateError{
			StatusCode: flux.StatusServerError,
			Message:    "PERMISSION:PROVIDER:UNKNOWN_PROTOCOL",
			Internal:   err,
		}
	}
	provideEndpoint := &flux.Endpoint{
		UpstreamHost:   meta.UpstreamHost,
		UpstreamMethod: meta.UpstreamMethod,
		UpstreamUri:    meta.UpstreamUri,
		Arguments:      meta.Arguments,
	}
	if ret, err := provider.Invoke(provideEndpoint, ctx); nil != err {
		logger.Trace(ctx.RequestId()).Errorw("Permission Provider backend load error",
			"provider-proto", meta.UpstreamProto, "provider-uri", meta.UpstreamUri, "provider-method", meta.UpstreamMethod, "error", err)
		return false, cache.NoExpiration, &flux.StateError{
			StatusCode: flux.StatusServerError,
			Message:    "PERMISSION:PROVIDER:LOAD",
			Internal:   err,
		}
	} else {
		passed, expire, err := GetEndpointPermissionResponseDecoder()(ret, ctx)
		if nil != err {
			logger.Trace(ctx.RequestId()).Errorw("Permission decode response error",
				"provider-proto", meta.UpstreamProto, "provider-uri", meta.UpstreamUri, "provider-method", meta.UpstreamMethod, "error", err)
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
	logger.Trace(ctx.RequestId()).Infow("Decode endpoint permission",
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

func _newArgumentsKey(args []flux.Argument) string {
	// [(T:v1),(T:v2),]
	sb := new(strings.Builder)
	sb.WriteByte('[')
	for _, arg := range args {
		sb.WriteString(_newArgumentKey(arg))
		sb.WriteByte(',')
	}
	sb.WriteByte(']')
	return sb.String()
}

func _newArgumentKey(arg flux.Argument) string {
	// (T:val)
	sb := new(strings.Builder)
	sb.WriteByte('(')
	sb.WriteString(arg.TypeClass)
	sb.WriteByte(':')
	if flux.ArgumentTypeComplex == arg.Type && len(arg.Fields) > 0 {
		sb.WriteString(_newArgumentsKey(arg.Fields))
	} else {
		sb.WriteString(cast.ToString(arg.HttpValue.Value()))
	}
	sb.WriteByte(')')
	return sb.String()
}
