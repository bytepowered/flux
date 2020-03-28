package filter

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/pkg"
	"github.com/bytepowered/lakego"
	"strings"
	"time"
)

const (
	FilterIdPermissionVerification = "PermissionVerificationFilter"
)

var (
	ErrPermissionSubjectNotFound = &flux.InvokeError{
		StatusCode: flux.StatusBadRequest,
		Message:    "PERMISSION:SUBJECT_NOT_FOUND",
	}
	ErrPermissionSubjectAccessDenied = &flux.InvokeError{
		StatusCode: flux.StatusAccessDenied,
		Message:    "PERMISSION:SUBJECT_ACCESS_DENIED",
	}
)

type PermissionProvider func(subjectId, method, uri string) (bool, error)

func PermissionVerificationFactory() interface{} {
	return NewPermissionVerificationWith(nil)
}

func NewPermissionVerificationWith(provider PermissionProvider) flux.Filter {
	return &PermissionVerificationFilter{
		provider: provider,
	}
}

// Permission Filter，负责读取JWT的Subject字段，调用指定Dubbo接口判断权限
type PermissionVerificationFilter struct {
	disabled  bool
	provider  PermissionProvider
	permCache lakego.Cache
}

func (p *PermissionVerificationFilter) Invoke(next flux.FilterInvoker) flux.FilterInvoker {
	if p.disabled {
		return next
	}
	return func(ctx flux.Context) *flux.InvokeError {
		if false == ctx.Endpoint().Authorize {
			return next(ctx)
		}
		if err := p.verify(ctx); nil != err {
			return err
		} else {
			return next(ctx)
		}
	}
}

func (p *PermissionVerificationFilter) Init(config flux.Config) error {
	p.disabled = config.BooleanOrDefault(keyConfigDisabled, false)
	if p.disabled {
		logger.Infof("Permission filter was DISABLED!!")
		return nil
	}
	if !config.IsEmpty() && p.provider == nil {
		p.provider = func() PermissionProvider {
			proto := config.String(keyConfigVerificationProtocol)
			upsHost := config.String(keyConfigVerificationHost)
			upsUri := config.String(keyConfigVerificationUri)
			upsMethod := config.String(keyConfigVerificationMethod)
			logger.Infof("Permission filter config provider, proto:%s, method: %s, uri: %s%s", proto, upsMethod, upsHost, upsHost)
			return func(subjectId, method, pattern string) (bool, error) {
				switch strings.ToUpper(proto) {
				case flux.ProtocolDubbo:
					return _loadPermByExchange(flux.ProtocolDubbo, upsHost, upsMethod, upsUri, subjectId, method, pattern)
				case flux.ProtocolHttp:
					return _loadPermByExchange(flux.ProtocolHttp, upsHost, upsMethod, upsUri, subjectId, method, pattern)
				default:
					return false, fmt.Errorf("unknown verification protocol: %s", proto)
				}
			}
		}()
	}
	logger.Infof("Permission filter initializing, config: %+v", config)
	// TODO 检查参数
	permCacheExpiration := config.Int64OrDefault(keyConfigCacheExpiration, defValueCacheExpiration)
	p.permCache = lakego.NewSimple(lakego.WithExpiration(time.Minute * time.Duration(permCacheExpiration)))
	return nil
}

func (p *PermissionVerificationFilter) Order() int {
	return OrderFilterPermissionVerification
}

func (*PermissionVerificationFilter) Id() string {
	return FilterIdPermissionVerification
}

func (p *PermissionVerificationFilter) verify(ctx flux.Context) *flux.InvokeError {
	jwtSubjectId, ok := ctx.AttrValue(flux.XJwtSubject)
	if !ok {
		return ErrPermissionSubjectNotFound
	}
	endpoint := ctx.Endpoint()
	// 验证用户是否有权限访问API：(userSubId, method, uri-pattern)
	permKey := fmt.Sprintf("%s@%s#%s", jwtSubjectId, endpoint.HttpMethod, endpoint.HttpPattern)
	allowed, err := p.permCache.GetOrLoad(permKey, func(_ lakego.Key) (lakego.Value, error) {
		strSubId := pkg.ToString(jwtSubjectId)
		return p.provider(strSubId, endpoint.HttpMethod, endpoint.HttpPattern)
	})
	if err == nil {
		if v, ok := allowed.(bool); ok && v {
			return nil
		} else {
			return ErrPermissionSubjectAccessDenied
		}
	} else {
		return &flux.InvokeError{
			StatusCode: flux.StatusServerError,
			Message:    "PERMISSION:LOAD_ACCESS",
			Internal:   err,
		}
	}
}

func _loadPermByExchange(proto string,
	upsHost, upsMethod, upsUri string,
	reqSubjectId, reqMethod, reqPattern string) (bool, error) {
	exchange, _ := ext.GetExchange(proto)
	if ret, err := exchange.Invoke(&flux.Endpoint{
		UpstreamHost:   upsHost,
		UpstreamMethod: upsMethod,
		UpstreamUri:    upsUri,
		Arguments: []flux.Argument{
			{TypeClass: pkg.JavaLangStringClassName, ArgName: "subjectId", ArgValue: flux.NewWrapValue(reqSubjectId)},
			{TypeClass: pkg.JavaLangStringClassName, ArgName: "method", ArgValue: flux.NewWrapValue(reqMethod)},
			{TypeClass: pkg.JavaLangStringClassName, ArgName: "pattern", ArgValue: flux.NewWrapValue(reqPattern)},
		},
	}, nil); nil != err {
		return false, err
	} else {
		return strings.Contains(pkg.ToString(ret), "success"), nil
	}
}
