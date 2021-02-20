package filter

import (
	"context"
	"errors"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/pkg"
	"net/http"
	"sync"
	"time"
)

const (
	HystrixConfigKeyTimeout                = "hystrix_timeout"
	HystrixConfigKeyMaxRequest             = "hystrix_max_requests"
	HystrixConfigKeyRequestVolumeThreshold = "hystrix_request_volume_threshold"
	HystrixConfigKeySleepWindow            = "hystrix_sleep_window"
	HystrixConfigKeyErrorPercentThreshold  = "hystrix_error_threshold"
)

const (
	TypeIdHystrixFilter = "hystrix_filter"
)

func NewHystrixFilter(c HystrixConfig) *HystrixFilter {
	return &HystrixFilter{
		Config: c,
		marks:  sync.Map{},
	}
}

type (
	// HystrixServiceNameFunc 用于构建服务标识的函数
	HystrixServiceNameFunc func(ctx flux.Context) (serviceName string)
	// HystrixDowngradeFunc 熔断降级处理函数
	HystrixDowngradeFunc func(ctx flux.Context) *flux.ServeError
)

// HystrixConfig 熔断器配置
type HystrixConfig struct {
	ServiceSkipFunc        flux.FilterSkipper
	ServiceNameFunc        HystrixServiceNameFunc
	ServiceDowngradeFunc   HystrixDowngradeFunc
	timeout                int
	maxConcurrentRequests  int
	requestVolumeThreshold int
	sleepWindow            int
	errorPercentThreshold  int
}

// HystrixFilter
type HystrixFilter struct {
	Config HystrixConfig
	marks  sync.Map
}

func (r *HystrixFilter) Init(config *flux.Configuration) error {
	logger.Info("Hystrix filter initializing")
	config.SetDefaults(map[string]interface{}{
		HystrixConfigKeyRequestVolumeThreshold: 20,
		HystrixConfigKeyErrorPercentThreshold:  50,
		HystrixConfigKeySleepWindow:            500,
		HystrixConfigKeyMaxRequest:             10,
		HystrixConfigKeyTimeout:                1000,
	})
	r.Config.timeout = int(config.GetInt64(HystrixConfigKeyTimeout))
	r.Config.maxConcurrentRequests = int(config.GetInt64(HystrixConfigKeyMaxRequest))
	r.Config.requestVolumeThreshold = int(config.GetInt64(HystrixConfigKeyRequestVolumeThreshold))
	r.Config.sleepWindow = int(config.GetInt64(HystrixConfigKeySleepWindow))
	r.Config.errorPercentThreshold = int(config.GetInt64(HystrixConfigKeyErrorPercentThreshold))
	// 检查必要配置
	if pkg.IsNil(r.Config.ServiceNameFunc) {
		return errors.New("Hystrix.ServiceNameFunc is nil")
	}
	// 默认实现
	if r.Config.ServiceSkipFunc == nil {
		r.Config.ServiceSkipFunc = func(c flux.Context) bool {
			return false
		}
	}
	if r.Config.ServiceDowngradeFunc == nil {
		r.Config.ServiceDowngradeFunc = DefaultDowngradeFunc
	}
	logger.Infow("Hystrix config",
		"timeout(ms)", r.Config.timeout,
		"max-concurrent-requests", r.Config.maxConcurrentRequests,
		"request-volume-threshold", r.Config.requestVolumeThreshold,
		"sleep-window(ms)", r.Config.sleepWindow,
		"error-percent-threshold", r.Config.errorPercentThreshold,
	)
	return nil
}

func (r *HystrixFilter) DoFilter(next flux.FilterHandler) flux.FilterHandler {
	return func(ctx flux.Context) *flux.ServeError {
		if r.Config.ServiceSkipFunc(ctx) {
			return next(ctx)
		}
		serviceName := r.Config.ServiceNameFunc(ctx)
		r.initCommand(serviceName)
		// check circuit
		work := func(_ context.Context) error {
			ctx.AddMetric("M-"+r.TypeId(), time.Since(ctx.StartAt()))
			return next(ctx)
		}
		var reterr *flux.ServeError
		fallback := func(_ context.Context, err error) error {
			// 返回两种类型Error：
			// 1. 执行 next() 返回 *ServeError；
			// 2. 熔断返回 hystrix.CircuitError;
			if serr, ok := err.(*flux.ServeError); ok {
				reterr = serr
			} else if cerr, ok := err.(hystrix.CircuitError); ok {
				logger.Infow("HYSTRIX:CIRCUITED/DOWNGRADE",
					"is-circuited", ok, "service-name", serviceName, "circuit-error", cerr)
				reterr = r.Config.ServiceDowngradeFunc(ctx)
			} else {
				reterr = &flux.ServeError{
					StatusCode: flux.StatusServerError,
					ErrorCode:  flux.ErrorCodeGatewayInternal,
					Message:    "CIRCUIT:UNEXPECTED_ERROR",
					Internal:   err,
				}
			}
			return nil // fallback dont return errors
		}
		_ = hystrix.DoC(ctx.Context(), serviceName, work, fallback)
		return reterr
	}
}

func DefaultDowngradeFunc(ctx flux.Context) *flux.ServeError {
	return &flux.ServeError{
		StatusCode: http.StatusServiceUnavailable,
		ErrorCode:  flux.ErrorCodeGatewayCircuited,
		Message:    "Server busy",
	}
}

func (r *HystrixFilter) initCommand(serviceName string) {
	if _, exist := r.marks.LoadOrStore(serviceName, true); !exist {
		logger.Infow("HYSTRIX:COMMAND:INIT", "service-name", serviceName)
		hystrix.ConfigureCommand(serviceName, hystrix.CommandConfig{
			Timeout:                r.Config.timeout,
			MaxConcurrentRequests:  r.Config.maxConcurrentRequests,
			SleepWindow:            r.Config.sleepWindow,
			ErrorPercentThreshold:  r.Config.errorPercentThreshold,
			RequestVolumeThreshold: r.Config.requestVolumeThreshold,
		})
	}
}

func (*HystrixFilter) TypeId() string {
	return TypeIdHystrixFilter
}
