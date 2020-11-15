package filter

import (
	"errors"
	"fmt"
	"github.com/bytepowered/cache"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/pkg"
	"github.com/spf13/cast"
	"reflect"
	"strings"
	"time"
)

const (
	TypeIdPermissionFilter = "PermissionFilter"
)

func init() {
	SetPermissionResponseDecoder(defaultPermissionDecoder)
	SetPermissionKeyFunc(defaultPermissionKey)
}

var (
	_permissionResponseDecoder PermissionResponseDecoder
	_permissionKeyFunc         PermissionKeyFunc
)

type (
	// PermissionKeyFunc 生成权限Key的函数
	PermissionKeyFunc func(ctx flux.Context) (key string, err error)
	// ResponseDecoder 权限验证结果解析函数
	PermissionResponseDecoder func(response interface{}, ctx flux.Context) (pass bool, expire time.Duration, err error)
)

func PermissionFilterFactory() interface{} {
	return &PermissionFilter{}
}

// PermissionFilter 提供基于Endpoint.Permission元数据的权限验证
type PermissionFilter struct {
	disabled           bool
	permissions        cache.Cache
	ResponseDecoder    PermissionResponseDecoder
	PermissionSkipFunc flux.FilterSkipper
	PermissionKeyFunc  PermissionKeyFunc
}

func (p *PermissionFilter) Init(config *flux.Configuration) error {
	config.SetDefaults(map[string]interface{}{
		ConfigKeyCacheExpiration: DefaultValueCacheExpiration,
		ConfigKeyCacheSize:       DefaultValueCacheSize,
		ConfigKeyDisabled:        false,
	})
	p.disabled = config.GetBool(ConfigKeyDisabled)
	if p.disabled {
		logger.Info("Endpoint permission filter was DISABLED!!")
		return nil
	}
	expiration := config.GetDuration(ConfigKeyCacheExpiration)
	size := config.GetInt(ConfigKeyCacheSize)
	p.permissions = cache.New(size).Expiration(expiration).LRU().Build()
	logger.Infow("Endpoint permission filter init", "cache-alg", "ExpireLRU", "cache-size", size, "cache-expire", expiration.String())
	if pkg.IsNil(p.PermissionKeyFunc) {
		p.PermissionKeyFunc = GetPermissionKeyFunc()
	}
	if pkg.IsNil(p.ResponseDecoder) {
		p.ResponseDecoder = GetPermissionResponseDecoder()
	}
	if pkg.IsNil(p.PermissionSkipFunc) {
		p.PermissionSkipFunc = func(_ flux.Context) bool {
			return false
		}
	}
	return nil
}

func (p *PermissionFilter) DoFilter(next flux.FilterHandler) flux.FilterHandler {
	if p.disabled {
		return next
	}
	return func(ctx flux.Context) *flux.StateError {
		// 必须开启Authorize才进行权限校验
		endpoint := ctx.Endpoint()
		provider := endpoint.Permission
		if false == endpoint.Authorize || !provider.IsValid() {
			return next(ctx)
		}
		if p.PermissionSkipFunc(ctx) {
			return next(ctx)
		}
		toStateError := func(err error, msg string) *flux.StateError {
			if serr, ok := err.(*flux.StateError); ok {
				return serr
			} else {
				return &flux.StateError{
					StatusCode: flux.StatusServerError,
					ErrorCode:  msg,
					Message:    msg,
					Internal:   err,
				}
			}
		}
		// 权限验证结果缓存
		permissionKey, err := p.PermissionKeyFunc(ctx)
		if nil != err {
			return toStateError(err, "PERMISSION:GENERATE:KEY")
		}
		passed, err := p.permissions.GetOrLoad(permissionKey, func(_ interface{}) (interface{}, *time.Duration, error) {
			return p.doPermissionVerify(&provider, ctx)
		})
		servInface, servMethod := ctx.ServiceName()
		serviceName := servInface + "." + servMethod
		providerName := provider.Interface + "." + provider.Method
		logger.TraceContext(ctx).Infow("Permission verified",
			"target-service", serviceName, "passed", passed, "permission-key", permissionKey, "provider", providerName)
		if nil != err {
			return toStateError(err, "PERMISSION:LOAD:ERROR")
		}
		if true == cast.ToBool(passed) {
			return next(ctx)
		} else {
			return err.(*flux.StateError)
		}
	}
}

func (*PermissionFilter) TypeId() string {
	return TypeIdPermissionFilter
}

func (p *PermissionFilter) doPermissionVerify(provider *flux.PermissionService, ctx flux.Context) (pass bool, expire *time.Duration, err *flux.StateError) {
	backend, ok := ext.GetBackend(provider.RpcProto)
	if !ok {
		logger.TraceContext(ctx).Errorw("Permission provider backend unsupported protocol",
			"provider-proto", provider.RpcProto, "provider-uri", provider.Interface, "provider-method", provider.Method)
		return false, cache.NoExpiration, &flux.StateError{
			StatusCode: flux.StatusServerError,
			Message:    "PERMISSION:PROVIDER:UNKNOWN_PROTOCOL",
			Internal:   err,
		}
	}
	// Invoke to check permission
	if ret, err := backend.Invoke(flux.BackendService{
		RemoteHost: provider.RemoteHost,
		Method:     provider.Method,
		Interface:  provider.Interface,
		Arguments:  provider.Arguments,
	}, ctx); nil != err {
		logger.TraceContext(ctx).Errorw("Permission provider backend load error",
			"provider-proto", provider.RpcProto, "provider-uri", provider.Interface, "provider-method", provider.Method, "error", err)
		return false, cache.NoExpiration, &flux.StateError{
			StatusCode: flux.StatusServerError,
			Message:    "PERMISSION:PROVIDER:LOAD",
			Internal:   err,
		}
	} else {
		passed, expire, err := p.ResponseDecoder(ret, ctx)
		if nil != err {
			logger.TraceContext(ctx).Errorw("Permission decode response error",
				"provider-proto", provider.RpcProto, "provider-uri", provider.Interface, "provider-method", provider.Method, "error", err)
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

func SetPermissionResponseDecoder(decoder PermissionResponseDecoder) {
	_permissionResponseDecoder = decoder
}

func GetPermissionResponseDecoder() PermissionResponseDecoder {
	return _permissionResponseDecoder
}

func SetPermissionKeyFunc(f PermissionKeyFunc) {
	_permissionKeyFunc = f
}

func GetPermissionKeyFunc() PermissionKeyFunc {
	return _permissionKeyFunc
}

func defaultPermissionDecoder(response interface{}, ctx flux.Context) (bool, time.Duration, error) {
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

// defaultPermissionKey 默认生成权限Key
func defaultPermissionKey(ctx flux.Context) (string, error) {
	permission := ctx.Endpoint().Permission
	lookup := ext.GetArgumentValueLookupFunc()
	resolver := ext.GetArgumentValueResolveFunc()
	argsKey, err := _newArgumentsKey(permission.Arguments, lookup, resolver, ctx)
	if nil != err {
		return "", err
	}
	// 以Permission的(ServiceTag + 具体参数Value列表)来构建单个请求的缓存Key
	serviceName := flux.NewServiceKey(permission.RpcProto, permission.RemoteHost, permission.Method, permission.Interface)
	return serviceName + "#" + argsKey, nil
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
