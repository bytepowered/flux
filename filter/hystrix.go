package filter

import (
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

type HystrixConfig struct {
	Timeout                int
	MaxConcurrentRequests  int
	RequestVolumeThreshold int
	SleepWindow            int
	ErrorPercentThreshold  int
}

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
	r.config = &HystrixConfig{
		Timeout:                int(config.GetInt64(HystrixConfigKeyTimeout)),
		MaxConcurrentRequests:  int(config.GetInt64(HystrixConfigKeyMaxRequest)),
		RequestVolumeThreshold: int(config.GetInt64(HystrixConfigKeyRequestVolumeThreshold)),
		SleepWindow:            int(config.GetInt64(HystrixConfigKeySleepWindow)),
		ErrorPercentThreshold:  int(config.GetInt64(HystrixConfigKeyErrorPercentThreshold)),
	}
	return nil
}

func (r *HystrixFilter) Invoke(next flux.FilterInvoker) flux.FilterInvoker {
	return func(ctx flux.Context) *flux.StateError {
		// 只处理Http协议，Dubbo协议内部自带熔断逻辑
		ep := ctx.Endpoint()
		if flux.ProtoHttp != ep.Protocol {
			return next(ctx)
		}
		// Proto/Host/Uri 可以标识一个服务。Host可能为空，直接在Url中展示
		serviceKey := fmt.Sprintf("%s:%s/%s", ep.Protocol, ep.UpstreamHost, ep.UpstreamUri)
		r.initCommand(serviceKey)
		err := hystrix.Do(serviceKey, func() error {
			return next(ctx)
		}, func(err error) error {
			_, ok := err.(hystrix.CircuitError)
			logger.Trace(ctx.RequestId()).Debugw("Hystrix check", "ok", ok, "service", serviceKey, "error", err)
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

func (r *HystrixFilter) initCommand(serviceKey string) {
	if _, exist := r.marks.LoadOrStore(serviceKey, true); !exist {
		hystrix.ConfigureCommand(serviceKey, hystrix.CommandConfig{
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
