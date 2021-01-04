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
)

const (
	HystrixConfigKeyTimeout                = "hystrix-timeout"
	HystrixConfigKeyMaxRequest             = "hystrix-max-requests"
	HystrixConfigKeyRequestVolumeThreshold = "hystrix-request-volume-threshold"
	HystrixConfigKeySleepWindow            = "hystrix-sleep-window"
	HystrixConfigKeyErrorPercentThreshold  = "hystrix-error-threshold"
)

const (
	TypeIdHystrixFilter = "HystrixFilter"
)

func NewHystrixFilter(c HystrixConfig) flux.Filter {
	return &HystrixFilter{
		Config: c,
		marks:  sync.Map{},
	}
}

type (
	// HystrixServiceNameFunc 用于构建服务标识的函数
	HystrixServiceNameFunc func(ctx flux.Context) (serviceName string)
	// HystrixServiceTestFunc 用于测试StateError是否需要熔断
	HystrixServiceTestFunc func(err *flux.ServeError) (circuited bool)
)

// HystrixConfig
type HystrixConfig struct {
	ServiceSkipFunc        flux.FilterSkipper
	ServiceNameFunc        HystrixServiceNameFunc
	ServiceTestFunc        HystrixServiceTestFunc
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
	if pkg.IsNil(r.Config.ServiceSkipFunc) {
		r.Config.ServiceSkipFunc = func(c flux.Context) bool {
			return false
		}
	}
	if pkg.IsNil(r.Config.ServiceNameFunc) {
		return errors.New("Hystrix.ServiceNameFunc is nil")
	}
	if pkg.IsNil(r.Config.ServiceTestFunc) {
		return errors.New("Hystrix.ServiceTestFunc is nil")
	}
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
		err := hystrix.DoC(ctx.Context(), serviceName, func(_ context.Context) error {
			ctx.AddMetric("M-"+r.TypeId(), ctx.ElapsedTime())
			if ierr := next(ctx); nil != ierr && r.Config.ServiceTestFunc(ierr) {
				return hystrix.CircuitError{Message: ierr.Message}
			} else {
				return nil
			}
		}, func(_ context.Context, err error) error {
			_, ok := err.(hystrix.CircuitError)
			logger.TraceContext(ctx).Infow("Hystrix check",
				"is-circuit-error", ok, "service-name", serviceName, "error", err)
			return err
		})
		if nil == err {
			return nil
		}
		msg := flux.ErrorMessageHystrixCircuited
		if ce, ok := err.(hystrix.CircuitError); ok {
			msg = ce.Message
		}
		return &flux.ServeError{
			StatusCode: http.StatusBadGateway,
			ErrorCode:  flux.ErrorCodeGatewayCircuited,
			Message:    msg,
			Internal:   err,
		}
	}
}

func (r *HystrixFilter) initCommand(serviceName string) {
	if _, exist := r.marks.LoadOrStore(serviceName, true); !exist {
		logger.Infof("Hystrix create command", "service-name", serviceName)
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
