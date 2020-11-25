package buildin

import (
	"fmt"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/ext"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/pkg"
	"net/http"
)

const (
	TypeIdPermissionV2Filter = "PermissionV2Filter"
)

const (
	ErrorCodePermissionDenied = "PERMISSION:ACCESS_DENIED"
)

type (
	// PermissionVerifyReport 权限验证结果报告
	PermissionVerifyReport struct {
		StatusCode int    `json:"statusCode"`
		Success    bool   `json:"success"`
		ErrorCode  string `json:"errorCode"`
		Message    string `json:"message"`
	}
	// PermissionVerifyFunc 权限验证
	// @return pass 对当前请求的权限验证是否通过；
	// @return err 如果验证过程发生错误，返回error；
	PermissionVerifyFunc func(ctx flux.Context) (report PermissionVerifyReport, err error)
)

// PermissionV2Config 权限配置
type PermissionV2Config struct {
	SkipFunc   flux.FilterSkipper
	VerifyFunc PermissionVerifyFunc
}

func NewPermissionVerifyReport(success bool, errorCode, message string) PermissionVerifyReport {
	return PermissionVerifyReport{
		StatusCode: flux.StatusUnauthorized,
		Success:    success,
		ErrorCode:  errorCode,
		Message:    message,
	}
}

func NewPermissionV2Filter(c PermissionV2Config) *PermissionV2Filter {
	return &PermissionV2Filter{
		Configs: c,
	}
}

// PermissionV2Filter 提供基于Endpoint.Permission元数据的权限验证
type PermissionV2Filter struct {
	Disabled bool
	Configs  PermissionV2Config
}

func (p *PermissionV2Filter) Init(config *flux.Configuration) error {
	config.SetDefaults(map[string]interface{}{
		ConfigKeyDisabled: false,
	})
	p.Disabled = config.GetBool(ConfigKeyDisabled)
	if p.Disabled {
		logger.Info("Endpoint PermissionV2Filter was DISABLED!!")
		return nil
	}
	if pkg.IsNil(p.Configs.SkipFunc) {
		p.Configs.SkipFunc = func(_ flux.Context) bool {
			return false
		}
	}
	if pkg.IsNil(p.Configs.VerifyFunc) {
		return fmt.Errorf("PermissionV2Filter.VerifyFunc is nil")
	}
	return nil
}

func (*PermissionV2Filter) TypeId() string {
	return TypeIdPermissionV2Filter
}

func (p *PermissionV2Filter) DoFilter(next flux.FilterHandler) flux.FilterHandler {
	if p.Disabled {
		return next
	}
	return func(ctx flux.Context) *flux.StateError {
		if p.Configs.SkipFunc(ctx) {
			return next(ctx)
		}
		// 必须开启Authorize才进行权限校验
		endpoint := ctx.Endpoint()
		permission := endpoint.Permission
		if false == endpoint.Authorize || !permission.IsValid() {
			return next(ctx)
		}
		report, err := p.Configs.VerifyFunc(ctx)
		if nil != err {
			if serr, ok := err.(*flux.StateError); ok {
				return serr
			}
			return &flux.StateError{
				StatusCode: http.StatusForbidden,
				ErrorCode:  flux.ErrorCodeGatewayInternal,
				Message:    "PERMISSION:VERIFY:ERROR",
				Internal:   err,
			}
		}
		if !report.Success {
			return &flux.StateError{
				StatusCode: report.StatusCode,
				ErrorCode:  report.ErrorCode,
				Message:    report.Message,
			}
		}
		return next(ctx)
	}
}

// InvokeService 执行权限验证的后端服务，获取响应结果；
func (p *PermissionV2Filter) InvokeService(permission flux.PermissionService, ctx flux.Context) (interface{}, *flux.StateError) {
	backend, ok := ext.GetBackend(permission.RpcProto)
	if !ok {
		return nil, &flux.StateError{
			StatusCode: flux.StatusServerError,
			ErrorCode:  flux.ErrorCodeGatewayInternal,
			Message:    "PERMISSION:UNKNOWN_PROTOCOL",
			Internal:   fmt.Errorf("unknown protocol:%s", permission.RpcProto),
		}
	}
	// Invoke to check permission
	resp, err := backend.Invoke(flux.BackendService{
		RemoteHost: permission.RemoteHost,
		Method:     permission.Method,
		Interface:  permission.Interface,
		Arguments:  permission.Arguments,
	}, ctx)
	if nil != err {
		return nil, err
	} else {
		return resp, nil
	}
}
