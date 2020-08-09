package filter

import (
	"fmt"
	"github.com/bytepowered/cache"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/spf13/cast"
	"strings"
	"time"
)

const (
	TypeIdPermissionVerification = "PermissionVerificationFilter"
)

var (
	ErrPermissionSubjectNotFound = &flux.StateError{
		StatusCode: flux.StatusBadRequest,
		ErrorCode:  flux.ErrorCodeRequestInvalid,
		Message:    "PERMISSION:SUBJECT_NOT_FOUND",
	}
	ErrPermissionSubjectAccessDenied = &flux.StateError{
		StatusCode: flux.StatusAccessDenied,
		ErrorCode:  flux.ErrorCodeRequestInvalid,
		Message:    "PERMISSION:SUBJECT_ACCESS_DENIED",
	}
)

// 权限验证函数 检查SubjectId,Method,Patter是否具有权限
type PermissionVerificationFunc func(subjectId, method, pattern string) (bool, *time.Duration, error)

func PermissionVerificationFactory() interface{} {
	return NewPermissionVerificationWith(nil)
}

func NewPermissionVerificationWith(provider PermissionVerificationFunc) flux.Filter {
	return &PermissionVerificationFilter{
		provider: provider,
	}
}

// Permission Filter，负责读取JWT的Subject字段，调用指定Dubbo接口判断权限
type PermissionVerificationFilter struct {
	disabled  bool
	provider  PermissionVerificationFunc
	permCache cache.Cache
}

func (p *PermissionVerificationFilter) Invoke(next flux.FilterInvoker) flux.FilterInvoker {
	if p.disabled {
		return next
	}
	return func(ctx flux.Context) *flux.StateError {
		if false == ctx.Endpoint().Authorize {
			return next(ctx)
		}
		if err := p.doVerification(ctx); nil != err {
			return err
		} else {
			return next(ctx)
		}
	}
}

func (p *PermissionVerificationFilter) Init(config *flux.Configuration) error {
	config.SetDefaults(map[string]interface{}{
		ConfigKeyCacheExpiration: defValueCacheExpiration,
		ConfigKeyCacheSize:       defValueCacheSize,
		ConfigKeyDisabled:        false,
	})
	p.disabled = config.GetBool(ConfigKeyDisabled)
	if p.disabled {
		logger.Infof("Permission filter was DISABLED!!")
		return nil
	}
	logger.Infof("Permission filter initializing")
	if config.IsSet(UpstreamConfigKeyProtocol, UpstreamConfigKeyUri, UpstreamConfigKeyMethod) && p.provider == nil {
		p.provider = func() PermissionVerificationFunc {
			proto := config.GetString(UpstreamConfigKeyProtocol)
			host := config.GetString(UpstreamConfigKeyHost)
			uri := config.GetString(UpstreamConfigKeyUri)
			method := config.GetString(UpstreamConfigKeyMethod)
			logger.Infof("Permission filter config provider, proto:%s, method: %s, uri: %s%s", proto, method, host, host)
			return func(subjectId, method, pattern string) (bool, *time.Duration, error) {
				switch strings.ToUpper(proto) {
				case flux.ProtoDubbo:
					return _loadPermByExchange(flux.ProtoDubbo, host, method, uri, subjectId, method, pattern)
				case flux.ProtoHttp:
					return _loadPermByExchange(flux.ProtoHttp, host, method, uri, subjectId, method, pattern)
				default:
					return false, cache.NoExpiration, fmt.Errorf("unknown verification protocol: %s", proto)
				}
			}
		}()
	}
	expiration := time.Minute * time.Duration(config.GetInt64(ConfigKeyCacheExpiration))
	size := config.GetInt(ConfigKeyCacheSize)
	p.permCache = cache.New(size).Expiration(expiration).LRU().Build()
	return nil
}

func (p *PermissionVerificationFilter) Order() int {
	return OrderFilterPermissionVerification
}

func (*PermissionVerificationFilter) TypeId() string {
	return TypeIdPermissionVerification
}

func (p *PermissionVerificationFilter) doVerification(ctx flux.Context) *flux.StateError {
	jwtSubjectId, ok := ctx.GetAttribute(flux.XJwtSubject)
	if !ok {
		return ErrPermissionSubjectNotFound
	}
	endpoint := ctx.Endpoint()
	// 验证用户是否有权限访问API：(userSubId, method, uri-pattern)
	permKey := fmt.Sprintf("%s@%s#%s", jwtSubjectId, endpoint.HttpMethod, endpoint.HttpPattern)
	allowed, err := p.permCache.GetOrLoad(permKey, func(key interface{}) (interface{}, *time.Duration, error) {
		strSubId := cast.ToString(jwtSubjectId)
		return p.provider(strSubId, endpoint.HttpMethod, endpoint.HttpPattern)
	})
	if err == nil {
		if v, ok := allowed.(bool); ok && v {
			return nil
		} else {
			return ErrPermissionSubjectAccessDenied
		}
	} else {
		return &flux.StateError{
			StatusCode: flux.StatusServerError,
			Message:    "PERMISSION:LOAD_ACCESS",
			Internal:   err,
		}
	}
}

func _loadPermByExchange(proto string, host, method, uri string, reqSubjectId, reqMethod, reqPattern string) (bool, *time.Duration, error) {
	exchange, _ := ext.GetExchange(proto)
	if ret, err := exchange.Invoke(&flux.Endpoint{
		UpstreamHost:   host,
		UpstreamMethod: method,
		UpstreamUri:    uri,
		Arguments: []flux.Argument{
			ext.NewStringArgument("subjectId", reqSubjectId),
			ext.NewStringArgument("method", reqMethod),
			ext.NewStringArgument("pattern", reqPattern),
		},
	}, nil); nil != err {
		return false, cache.NoExpiration, err
	} else {
		return strings.Contains(cast.ToString(ret), "success"), cache.NoExpiration, nil
	}
}
