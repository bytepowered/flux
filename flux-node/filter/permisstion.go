package filter

import (
	"errors"
	"fmt"
	flux2 "github.com/bytepowered/flux/flux-node"
	"github.com/bytepowered/flux/flux-node/backend"
	"github.com/bytepowered/flux/flux-node/ext"
	"github.com/bytepowered/flux/flux-node/logger"
	"github.com/bytepowered/flux/flux-pkg"
	"net/http"
	"time"
)

const (
	TypeIdPermissionV2Filter = "permission_filter"
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
	PermissionVerifyFunc func(services []flux2.BackendService, ctx flux2.Context) (report PermissionVerifyReport, err error)
)

// PermissionConfig 权限配置
type PermissionConfig struct {
	SkipFunc   flux2.FilterSkipper
	VerifyFunc PermissionVerifyFunc
}

func NewPermissionVerifyReport(success bool, errorCode, message string) PermissionVerifyReport {
	return PermissionVerifyReport{
		StatusCode: flux2.StatusUnauthorized,
		Success:    success,
		ErrorCode:  errorCode,
		Message:    message,
	}
}

func NewPermissionFilter(c PermissionConfig) *PermissionFilter {
	return &PermissionFilter{
		Configs: c,
	}
}

// PermissionFilter 提供基于Endpoint.Permission元数据的权限验证
type PermissionFilter struct {
	Disabled bool
	Configs  PermissionConfig
}

func (p *PermissionFilter) Init(config *flux2.Configuration) error {
	config.SetDefaults(map[string]interface{}{
		ConfigKeyDisabled: false,
	})
	p.Disabled = config.GetBool(ConfigKeyDisabled)
	if p.Disabled {
		logger.Info("Endpoint PermissionFilter was DISABLED!!")
		return nil
	}
	if fluxpkg.IsNil(p.Configs.SkipFunc) {
		p.Configs.SkipFunc = func(_ flux2.Context) bool {
			return false
		}
	}
	if fluxpkg.IsNil(p.Configs.VerifyFunc) {
		return fmt.Errorf("PermissionFilter.VerifyFunc is nil")
	}
	return nil
}

func (*PermissionFilter) TypeId() string {
	return TypeIdPermissionV2Filter
}

func (p *PermissionFilter) DoFilter(next flux2.FilterHandler) flux2.FilterHandler {
	if p.Disabled {
		return next
	}
	return func(ctx flux2.Context) *flux2.ServeError {
		if p.Configs.SkipFunc(ctx) {
			return next(ctx)
		}
		// 没有任何权限校验定义
		endpoint := ctx.Endpoint()
		size := len(endpoint.Permissions)
		if size == 0 && !endpoint.Permission.IsValid() {
			return next(ctx)
		}
		services := make([]flux2.BackendService, 0, 1+size)
		// Define permission first
		if endpoint.Permission.IsValid() {
			services = append(services, endpoint.Permission)
		}
		for _, id := range endpoint.Permissions {
			if srv, ok := ext.BackendServiceById(id); ok {
				services = append(services, srv)
			} else {
				return &flux2.ServeError{
					StatusCode: flux2.StatusServerError,
					ErrorCode:  flux2.ErrorCodeGatewayInternal,
					Message:    flux2.ErrorMessagePermissionServiceNotFound,
					CauseError: errors.New("permission.service not found, id: " + id),
				}
			}
		}
		report, err := p.Configs.VerifyFunc(services, ctx)
		ctx.AddMetric(p.TypeId(), time.Since(ctx.StartAt()))
		if nil != err {
			if serr, ok := err.(*flux2.ServeError); ok {
				return serr
			}
			return &flux2.ServeError{
				StatusCode: http.StatusForbidden,
				ErrorCode:  flux2.ErrorCodeGatewayInternal,
				Message:    flux2.ErrorMessagePermissionVerifyError,
				CauseError: err,
			}
		}
		if !report.Success {
			return &flux2.ServeError{
				StatusCode: EnsurePermissionStatusCode(report.StatusCode),
				ErrorCode:  EnsurePermissionErrorCode(report.ErrorCode),
				Message:    EnsurePermissionMessage(report.Message),
			}
		}
		return next(ctx)
	}
}

// InvokeCodec 执行权限验证的后端服务，获取响应结果；
func (p *PermissionFilter) InvokeCodec(ctx flux2.Context, service flux2.BackendService) (*flux2.BackendResponse, *flux2.ServeError) {
	return backend.DoInvokeCodec(ctx, service)
}

func EnsurePermissionStatusCode(status int) int {
	if status < 100 {
		return http.StatusForbidden
	}
	return status
}

func EnsurePermissionErrorCode(code string) string {
	if "" == code {
		return flux2.ErrorCodePermissionDenied
	}
	return code
}

func EnsurePermissionMessage(message string) string {
	if "" == message {
		return flux2.ErrorMessagePermissionAccessDenied
	}
	return message
}
