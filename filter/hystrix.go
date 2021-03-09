package filter

import (
	"context"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/pkg"
	"github.com/spf13/cast"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	HystrixConfigKeyTimeout               = "timeout"
	HystrixConfigKeyRequestMax            = "request_max"
	HystrixConfigKeyRequestThreshold      = "request_threshold"
	HystrixConfigKeyErrorPercentThreshold = "error_threshold"
	HystrixConfigKeySleepWindow           = "sleep_window"
	HystrixConfigApplication              = "applications"
	HystrixConfigService                  = "service"
)

const (
	TypeIdHystrixFilter = "hystrix_filter"
)

func NewHystrixFilter(c HystrixConfig) *HystrixFilter {
	return &HystrixFilter{
		HystrixConfig: c,
		commands:      sync.Map{},
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
	HystrixConfig
	commands     sync.Map
	services     *flux.Configuration
	applications *flux.Configuration
}

func (r *HystrixFilter) Init(c *flux.Configuration) error {
	logger.Info("Hystrix filter initializing")
	r.applications = c.Sub(HystrixConfigApplication)
	r.services = c.Sub(HystrixConfigService)
	c.SetDefaults(map[string]interface{}{
		HystrixConfigKeyRequestThreshold:      20,
		HystrixConfigKeyErrorPercentThreshold: 50,
		HystrixConfigKeySleepWindow:           1000,
		HystrixConfigKeyRequestMax:            200,
		HystrixConfigKeyTimeout:               10000,
	})
	r.HystrixConfig.timeout = int(c.GetInt64(HystrixConfigKeyTimeout))
	r.HystrixConfig.maxConcurrentRequests = int(c.GetInt64(HystrixConfigKeyRequestMax))
	r.HystrixConfig.requestVolumeThreshold = int(c.GetInt64(HystrixConfigKeyRequestThreshold))
	r.HystrixConfig.sleepWindow = int(c.GetInt64(HystrixConfigKeySleepWindow))
	r.HystrixConfig.errorPercentThreshold = int(c.GetInt64(HystrixConfigKeyErrorPercentThreshold))
	// 默认实现
	if pkg.IsNil(r.HystrixConfig.ServiceNameFunc) {
		r.HystrixConfig.ServiceNameFunc = func(ctx flux.Context) (name string) {
			return ctx.BackendServiceId()
		}
	}
	if r.HystrixConfig.ServiceSkipFunc == nil {
		r.HystrixConfig.ServiceSkipFunc = func(c flux.Context) bool {
			return false
		}
	}
	if r.HystrixConfig.ServiceDowngradeFunc == nil {
		r.HystrixConfig.ServiceDowngradeFunc = DefaultDowngradeFunc
	}
	logger.Infow("Hystrix default config",
		"timeout(ms)", r.HystrixConfig.timeout,
		"max-concurrent-requests", r.HystrixConfig.maxConcurrentRequests,
		"request-volume-threshold", r.HystrixConfig.requestVolumeThreshold,
		"sleep-window(ms)", r.HystrixConfig.sleepWindow,
		"error-percent-threshold", r.HystrixConfig.errorPercentThreshold,
	)
	return nil
}

func (r *HystrixFilter) DoFilter(next flux.FilterHandler) flux.FilterHandler {
	return func(ctx flux.Context) *flux.ServeError {
		if r.HystrixConfig.ServiceSkipFunc(ctx) {
			return next(ctx)
		}
		serviceName := r.HystrixConfig.ServiceNameFunc(ctx)
		r.initCommand(serviceName, ctx)
		// check circuit
		work := func(_ context.Context) error {
			ctx.AddMetric(r.TypeId(), time.Since(ctx.StartAt()))
			return next(ctx)
		}
		var reterr *flux.ServeError
		fallback := func(c context.Context, err error) error {
			// 返回两种类型Error：
			// 1. 执行 next() 返回 *ServeError；
			// 2. 熔断返回 hystrix.CircuitError;
			if serr, ok := err.(*flux.ServeError); ok {
				reterr = serr
			} else if cerr, ok := err.(hystrix.CircuitError); ok {
				logger.Infow("HYSTRIX:CIRCUITED/DOWNGRADE",
					"is-circuited", ok, "service-name", serviceName, "circuit-error", cerr)
				reterr = r.HystrixConfig.ServiceDowngradeFunc(ctx)
			} else if strings.Contains(err.Error(), context.Canceled.Error()) {
				reterr = &flux.ServeError{
					StatusCode: flux.StatusOK,
					ErrorCode:  flux.ErrorCodeGatewayCanceled,
					Message:    "CIRCUITED:CANCELED:BYCLIENT",
					CauseError: err,
				}
			} else {
				reterr = &flux.ServeError{
					StatusCode: flux.StatusServerError,
					ErrorCode:  flux.ErrorCodeGatewayInternal,
					Message:    "CIRCUITED:INTERNAL:UNEXPERR",
					CauseError: err,
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
		Message:    "CIRCUITED:SERVER_BUSY:DOWNGRADE",
	}
}

func (r *HystrixFilter) initCommand(serviceName string, ctx flux.Context) {
	if _, exist := r.commands.LoadOrStore(serviceName, true); !exist {
		logger.Infow("HYSTRIX:COMMAND:INIT", "service-name", serviceName)
		// 支持两种定制配置：
		// 1. 对单个服务接口配置；
		// 2. 对应用级别接口配置；
		conf := r.applications.Sub(ctx.Application())
		if r.services.IsSet(serviceName) {
			conf = r.services.Sub(serviceName)
		}
		hystrix.ConfigureCommand(serviceName, r.readConfig(conf, map[string]interface{}{
			HystrixConfigKeyTimeout:               r.HystrixConfig.timeout,
			HystrixConfigKeyRequestThreshold:      r.HystrixConfig.requestVolumeThreshold,
			HystrixConfigKeyRequestMax:            r.HystrixConfig.maxConcurrentRequests,
			HystrixConfigKeyErrorPercentThreshold: r.HystrixConfig.errorPercentThreshold,
			HystrixConfigKeySleepWindow:           r.HystrixConfig.sleepWindow,
		}))
	}
}

func (*HystrixFilter) TypeId() string {
	return TypeIdHystrixFilter
}

func (*HystrixFilter) readConfig(conf *flux.Configuration, defaults map[string]interface{}) hystrix.CommandConfig {
	getIntOr := func(k string) int {
		if conf.IsSet(k) {
			return int(conf.GetInt64(k))
		}
		return cast.ToInt(defaults[k])
	}
	return hystrix.CommandConfig{
		Timeout:                getIntOr(HystrixConfigKeyTimeout),
		MaxConcurrentRequests:  getIntOr(HystrixConfigKeyRequestMax),
		SleepWindow:            getIntOr(HystrixConfigKeyRequestThreshold),
		ErrorPercentThreshold:  getIntOr(HystrixConfigKeySleepWindow),
		RequestVolumeThreshold: getIntOr(HystrixConfigKeyErrorPercentThreshold),
	}
}
