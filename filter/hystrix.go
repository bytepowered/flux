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
	keyConfigHystrixTimeout                = "hystrix-timeout"
	keyConfigHystrixMaxRequest             = "hystrix-max-requests"
	keyConfigHystrixRequestVolumeThreshold = "hystrix-request-volume-threshold"
	keyConfigHystrixSleepWindow            = "hystrix-sleep-window"
	keyConfigHystrixErrorPercentThreshold  = "hystrix-error-threshold"
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
	Error                  []string
}

type HystrixFilter struct {
	config *HystrixConfig
	marks  sync.Map
}

func (r *HystrixFilter) Init(config flux.Configuration) error {
	r.config = &HystrixConfig{
		Timeout:                int(config.Int64OrDefault(keyConfigHystrixTimeout, 1000)),
		MaxConcurrentRequests:  int(config.Int64OrDefault(keyConfigHystrixMaxRequest, 10)),
		RequestVolumeThreshold: int(config.Int64OrDefault(keyConfigHystrixRequestVolumeThreshold, 20)),
		SleepWindow:            int(config.Int64OrDefault(keyConfigHystrixSleepWindow, 5000)),
		ErrorPercentThreshold:  int(config.Int64OrDefault(keyConfigHystrixErrorPercentThreshold, 50)),
	}
	logger.Infof("Hystrix filter initializing, config: %+v", r.config)
	return nil
}

func (r *HystrixFilter) Invoke(next flux.FilterInvoker) flux.FilterInvoker {
	return func(ctx flux.Context) *flux.InvokeError {
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
			logger.Debugf("Hystrix check, errors: %v, service: %v; %s", err, ok, serviceKey)
			return err
		})
		if nil == err {
			return nil
		}
		msg := "HYSTRIX:CIRCUITED"
		if ce, ok := err.(hystrix.CircuitError); ok {
			msg = ce.Message
		}
		return &flux.InvokeError{
			StatusCode: http.StatusBadGateway,
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
