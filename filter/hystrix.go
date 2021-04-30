package filter

import (
	"context"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
	"github.com/spf13/cast"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	ConfigKeyTimeout               = "timeout"
	ConfigKeyRequestMax            = "request_max"
	ConfigKeyRequestThreshold      = "request_threshold"
	ConfigKeyErrorPercentThreshold = "error_threshold"
	ConfigKeySleepWindow           = "sleep_window"
	ConfigApplication              = "applications"
	ConfigService                  = "service"
)

const (
	TypeIdHystrixFilter = "hystrix_filter"
)

const (
	hystrixRequestId = "hystrix.request.id"
)

func NewHystrixFilter(c HystrixConfig) *HystrixFilter {
	return &HystrixFilter{
		HystrixConfig: c,
		commands:      sync.Map{},
	}
}

type (
	// HystrixServiceNameFunc 用于构建服务标识的函数
	HystrixServiceNameFunc func(ctx *flux.Context) (serviceName string)
	// HystrixDowngradeFunc 熔断降级处理函数
	HystrixDowngradeFunc func(ctx *flux.Context, next flux.FilterInvoker, err error) *flux.ServeError
)

// HystrixConfig 熔断器配置
type HystrixConfig struct {
	ServiceSkipFunc      flux.FilterSkipper
	ServiceNameFunc      HystrixServiceNameFunc
	ServiceDowngradeFunc HystrixDowngradeFunc
	// globals
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
	r.applications = c.Sub(ConfigApplication)
	r.services = c.Sub(ConfigService)
	c.SetDefaults(map[string]interface{}{
		ConfigKeyRequestThreshold:      20,
		ConfigKeyErrorPercentThreshold: 50,
		ConfigKeyRequestMax:            1 * 1000,
		ConfigKeySleepWindow:           10 * 1000,
		ConfigKeyTimeout:               60 * 1000,
	})
	r.HystrixConfig.timeout = int(c.GetInt64(ConfigKeyTimeout))
	r.HystrixConfig.maxConcurrentRequests = int(c.GetInt64(ConfigKeyRequestMax))
	r.HystrixConfig.requestVolumeThreshold = int(c.GetInt64(ConfigKeyRequestThreshold))
	r.HystrixConfig.sleepWindow = int(c.GetInt64(ConfigKeySleepWindow))
	r.HystrixConfig.errorPercentThreshold = int(c.GetInt64(ConfigKeyErrorPercentThreshold))
	// 默认实现
	if r.HystrixConfig.ServiceNameFunc == nil {
		r.HystrixConfig.ServiceNameFunc = func(ctx *flux.Context) (name string) {
			return ctx.ServiceID()
		}
	}
	if r.HystrixConfig.ServiceSkipFunc == nil {
		r.HystrixConfig.ServiceSkipFunc = func(c *flux.Context) bool {
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

func (r *HystrixFilter) DoFilter(next flux.FilterInvoker) flux.FilterInvoker {
	return func(ctx *flux.Context) *flux.ServeError {
		if r.HystrixConfig.ServiceSkipFunc(ctx) {
			return next(ctx)
		}
		serviceName := r.HystrixConfig.ServiceNameFunc(ctx)
		r.initCommand(serviceName, ctx)
		// check circuit
		work := func(_ context.Context) error {
			ctx.AddMetric(r.FilterId(), time.Since(ctx.StartAt()))
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
				logger.Trace(c.Value(hystrixRequestId).(string)).Infow("HYSTRIX:CIRCUITED/DOWNGRADE",
					"is-circuited", ok, "service-name", serviceName, "circuit-error", cerr)
				reterr = r.HystrixConfig.ServiceDowngradeFunc(ctx, next, cerr)
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
		_ = hystrix.DoC(context.WithValue(ctx.Context(), hystrixRequestId, ctx.RequestId()), serviceName, work, fallback)
		return reterr
	}
}

func DefaultDowngradeFunc(ctx *flux.Context, next flux.FilterInvoker, err error) *flux.ServeError {
	return &flux.ServeError{
		StatusCode: http.StatusServiceUnavailable,
		ErrorCode:  flux.ErrorCodeGatewayCircuited,
		Message:    "CIRCUITED:SERVER_BUSY:DOWNGRADE",
		CauseError: err,
	}
}

func (r *HystrixFilter) initCommand(serviceName string, ctx *flux.Context) {
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
			ConfigKeyTimeout:               r.HystrixConfig.timeout,
			ConfigKeyRequestThreshold:      r.HystrixConfig.requestVolumeThreshold,
			ConfigKeyRequestMax:            r.HystrixConfig.maxConcurrentRequests,
			ConfigKeyErrorPercentThreshold: r.HystrixConfig.errorPercentThreshold,
			ConfigKeySleepWindow:           r.HystrixConfig.sleepWindow,
		}))
	}
}

func (*HystrixFilter) FilterId() string {
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
		Timeout:                getIntOr(ConfigKeyTimeout),
		MaxConcurrentRequests:  getIntOr(ConfigKeyRequestMax),
		RequestVolumeThreshold: getIntOr(ConfigKeyRequestThreshold),
		SleepWindow:            getIntOr(ConfigKeySleepWindow),
		ErrorPercentThreshold:  getIntOr(ConfigKeyErrorPercentThreshold),
	}
}
