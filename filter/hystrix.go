package filter

import (
	"context"
	"fmt"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
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

func HystrixFilterFactory() interface{} {
	return NewHystrixFilter()
}

func NewHystrixFilter() flux.Filter {
	return new(HystrixFilter)
}

type (
	// HystrixServiceNameFunc 用于构建服务标识的函数
	HystrixServiceNameFunc func(ctx flux.Context) (serviceName string)
	// HystrixServiceTestFunc 用于测试StateError是否需要熔断
	HystrixServiceTestFunc func(err *flux.StateError) (circuited bool)
)

// HystrixConfig
type HystrixConfig struct {
	Timeout                int
	MaxConcurrentRequests  int
	RequestVolumeThreshold int
	SleepWindow            int
	ErrorPercentThreshold  int
	ServiceSkipFunc        flux.FilterSkipper
	ServiceNameFunc        HystrixServiceNameFunc
	ServiceTestFunc        HystrixServiceTestFunc
}

// HystrixFilter
type HystrixFilter struct {
	config *HystrixConfig
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
	r.SetHystrixConfig(&HystrixConfig{
		Timeout:                int(config.GetInt64(HystrixConfigKeyTimeout)),
		MaxConcurrentRequests:  int(config.GetInt64(HystrixConfigKeyMaxRequest)),
		RequestVolumeThreshold: int(config.GetInt64(HystrixConfigKeyRequestVolumeThreshold)),
		SleepWindow:            int(config.GetInt64(HystrixConfigKeySleepWindow)),
		ErrorPercentThreshold:  int(config.GetInt64(HystrixConfigKeyErrorPercentThreshold)),
	})
	return nil
}

func (r *HystrixFilter) SetHystrixConfig(config *HystrixConfig) {
	r.config = config
	if r.config.ServiceSkipFunc == nil {
		r.config.ServiceSkipFunc = hystrixServiceSkipper
	}
	if r.config.ServiceNameFunc == nil {
		r.config.ServiceNameFunc = hystrixServiceNamer
	}
	if r.config.ServiceTestFunc == nil {
		r.config.ServiceTestFunc = hystrixServiceCircuited
	}
}

func (r *HystrixFilter) GetHystrixConfig() HystrixConfig {
	return *(r.config)
}

func (r *HystrixFilter) DoFilter(next flux.FilterHandler) flux.FilterHandler {
	return func(ctx flux.Context) *flux.StateError {
		if r.config.ServiceSkipFunc(ctx) {
			return next(ctx)
		}
		serviceName := r.config.ServiceNameFunc(ctx)
		r.initCommand(serviceName)
		// check circuit
		err := hystrix.DoC(ctx.Context(), serviceName, func(_ context.Context) error {
			if ierr := next(ctx); nil != ierr && r.config.ServiceTestFunc(ierr) {
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
		msg := "HYSTRIX:CIRCUITED"
		if ce, ok := err.(hystrix.CircuitError); ok {
			msg = ce.Message
		}
		return &flux.StateError{
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
			Timeout:                r.config.Timeout,
			MaxConcurrentRequests:  r.config.MaxConcurrentRequests,
			SleepWindow:            r.config.SleepWindow,
			ErrorPercentThreshold:  r.config.ErrorPercentThreshold,
			RequestVolumeThreshold: r.config.RequestVolumeThreshold,
		})
	}
}

func (*HystrixFilter) TypeId() string {
	return TypeIdHystrixFilter
}

// hystrixServiceSkipper 只处理Http协议，Dubbo协议内部自带熔断逻辑
func hystrixServiceSkipper(ctx flux.Context) bool {
	if flux.ProtoHttp != ctx.ServiceProto() {
		return true
	}
	return false
}

// hystrixServiceNamer 构建服务名称，Protocol/Host/Uri 可以标识一个服务。Host可能为空
func hystrixServiceNamer(ctx flux.Context) string {
	service := ctx.Endpoint().Service
	return fmt.Sprintf("%s:%s/%s", service.RpcProto, service.RemoteHost, service.Interface)
}

// hystrixServiceCircuited 判断是否熔断
func hystrixServiceCircuited(err *flux.StateError) bool {
	return nil != err
}
