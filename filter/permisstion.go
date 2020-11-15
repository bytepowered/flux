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
	"math"
	"net/http"
	"reflect"
	"strings"
	"time"
)

const (
	TypeIdPermissionFilter = "PermissionFilter"
)

func init() {
	SetPermissionResponseDecodeFunc(defaultPermissionResponseDecoder)
	SetPermissionKeyGenerateFunc(defaultPermissionGenerateKey)
}

var (
	_permissionResponseDecodeFunc PermissionResponseDecodeFunc
	_permissionKeyGenerateFunc    PermissionKeyGenerateFunc
)

type (
	// PermissionKeyGenerateFunc 生成权限Key的函数
	PermissionKeyGenerateFunc func(ctx flux.Context) (key string, err error)
	// PermissionVerifyReport 权限验证结果
	PermissionVerifyReport struct {
		AllowCache bool
		Passed     bool
		Expire     time.Duration
	}
	// ResponseDecodeFunc 权限验证结果解析函数
	PermissionResponseDecodeFunc func(response interface{}, ctx flux.Context) (report PermissionVerifyReport, err error)
)

func PermissionFilterFactory() interface{} {
	return &PermissionFilter{}
}

// PermissionFilter 提供基于Endpoint.Permission元数据的权限验证
type PermissionFilter struct {
	disabled           bool
	cacheDisabled      bool
	caching            cache.Cache
	PermissionSkipFunc flux.FilterSkipper
	ResponseDecodeFunc PermissionResponseDecodeFunc
	KeyGenerateFunc    PermissionKeyGenerateFunc
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
	p.cacheDisabled = config.GetBool(ConfigKeyCacheDisabled)
	if !p.cacheDisabled {
		p.caching = cache.New(size).LRU().Expiration(expiration).Build()
		logger.Infow("Endpoint permission filter init (use cached)", "cache-alg", "ExpireLRU", "cache-size", size, "cache-expire", expiration.String())
	} else {
		logger.Info("Endpoint permission filter init")
	}
	if pkg.IsNil(p.KeyGenerateFunc) {
		p.KeyGenerateFunc = GetPermissionKeyGenerateFunc()
	}
	if pkg.IsNil(p.ResponseDecodeFunc) {
		p.ResponseDecodeFunc = GetPermissionResponseDecodeFunc()
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
		loader := func(k interface{}) (interface{}, *time.Duration, error) {
			resp, err := p.doVerify(provider, ctx)
			if nil != err {
				return nil, nil, err
			}
			report, err := p.ResponseDecodeFunc(resp, ctx)
			if nil != err {
				return nil, nil, err
			}
			if report.AllowCache {
				return report.Passed, &report.Expire, nil
			} else {
				return report.Passed, cache.NoExpiration, nil
			}
		}
		var passed = false
		var err error = nil
		if p.cacheDisabled {
			if v, _, ex := loader(nil); nil != ex {
				err = ex
			} else {
				passed = cast.ToBool(v)
			}
		} else {
			if key, ex := p.KeyGenerateFunc(ctx); nil != ex {
				return &flux.StateError{
					StatusCode: flux.StatusServerError,
					ErrorCode:  "PERMISSION:GENERATE:KEY",
					Message:    "PERMISSION:" + ex.Error(),
					Internal:   ex,
				}
			} else if v, ex := p.caching.GetOrLoad(key, loader); nil != ex {
				err = ex
			} else {
				passed = cast.ToBool(v)
			}
		}
		if !pkg.IsNil(err) {
			return &flux.StateError{
				StatusCode: flux.StatusServerError,
				ErrorCode:  "PERMISSION:LOAD:ERROR",
				Message:    "PERMISSION:" + err.Error(),
				Internal:   err,
			}
		}
		if true != passed {
			return &flux.StateError{
				StatusCode: http.StatusForbidden,
				ErrorCode:  "PERMISSION:ACCESS_DENIED",
				Message:    "PERMISSION:VERIFY:ACCESS_DENIED",
			}
		}
		return next(ctx)
	}
}

func (*PermissionFilter) TypeId() string {
	return TypeIdPermissionFilter
}

func (p *PermissionFilter) doVerify(provider flux.PermissionService, ctx flux.Context) (response interface{}, err error) {
	backend, ok := ext.GetBackend(provider.RpcProto)
	if !ok {
		return nil, fmt.Errorf("provider unknown protocol:%s", provider.RpcProto)
	}
	// Invoke to check permission
	resp, err := backend.Invoke(flux.BackendService{
		RemoteHost: provider.RemoteHost,
		Method:     provider.Method,
		Interface:  provider.Interface,
		Arguments:  provider.Arguments,
	}, ctx)
	if nil != err {
		return nil, fmt.Errorf("provider load, error:%w", err)
	} else {
		return resp, nil
	}
}

func SetPermissionResponseDecodeFunc(decoder PermissionResponseDecodeFunc) {
	_permissionResponseDecodeFunc = decoder
}

func GetPermissionResponseDecodeFunc() PermissionResponseDecodeFunc {
	return _permissionResponseDecodeFunc
}

func SetPermissionKeyGenerateFunc(f PermissionKeyGenerateFunc) {
	_permissionKeyGenerateFunc = f
}

func GetPermissionKeyGenerateFunc() PermissionKeyGenerateFunc {
	return _permissionKeyGenerateFunc
}

func defaultPermissionResponseDecoder(response interface{}, ctx flux.Context) (PermissionVerifyReport, error) {
	logger.TraceContext(ctx).Infow("Decode endpoint permission",
		"response-type", reflect.TypeOf(response), "response", response)
	// 默认支持响应JSON数据：
	// {"status": "[success,error]", "permission": "[true,false]", "message": "OnErrorMessage", "expire": 5}
	strmap, ok := response.(map[string]interface{})
	if ok {
		if "success" == cast.ToString(strmap["status"]) {
			passed := cast.ToBool(strmap["permission"])
			nocache := cast.ToBool(strmap["nocache"])
			minutes := cast.ToInt(strmap["expire"])
			return PermissionVerifyReport{
				AllowCache: nocache,
				Expire:     time.Minute * time.Duration(math.Max(float64(minutes), 5)),
				Passed:     passed,
			}, nil
		} else {
			message := cast.ToString(strmap["message"])
			if "" == message {
				message = "Permission NOT SUCCESS, error message NOT FOUND"
			}
			return PermissionVerifyReport{
				AllowCache: false,
				Expire:     time.Duration(0),
				Passed:     false,
			}, errors.New(message)
		}
	}
	// 如果不是默认JSON结构的数据，只是包含success字符串，就是验证成功
	text := cast.ToString(response)
	return PermissionVerifyReport{
		AllowCache: strings.Contains(text, "success"),
		Expire:     time.Minute * 5,
		Passed:     false,
	}, nil
}

// defaultPermissionGenerateKey 默认生成权限Key
func defaultPermissionGenerateKey(ctx flux.Context) (string, error) {
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
